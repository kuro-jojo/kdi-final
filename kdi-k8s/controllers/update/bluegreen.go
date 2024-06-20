package update

import (
	"errors"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/utils"
)

var (
	servicesClient    corev1.ServiceInterface
	deploymentsClient appsv1.DeploymentInterface
)

// This file contains the Blue/Green strategie for updating a deployment

func UpdateUsingBlueGreenStrategy(c *gin.Context, updateForm UpdateForm) error {

	// Retrieve the namespace and deployment name from the URL parameters
	namespace := c.Param("namespace")
	deploymentName := c.Param("deployment")

	updateForm.Namespace = namespace
	updateForm.Name = deploymentName

	clientset := utils.GetClientSet(c)
	deploymentsClient = clientset.AppsV1().Deployments(updateForm.Namespace)
	servicesClient = clientset.CoreV1().Services(updateForm.Namespace)

	// Step 1: Retrieve the current deployment and service
	fmt.Println("Getting the current deployment")
	deployment, err := deploymentsClient.Get(c, updateForm.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get the current deployment: %v", err)
	}

	fmt.Println("Getting the associated service")
	service, err := getServiceByDeployment(c, deployment, updateForm.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get the associated service: %v", err)
	}

	// Step 2: Create the new deployment
	fmt.Println("Creating the new deployment")
	newDeployment, err := createNewDeployment(c, deployment, updateForm)
	if err != nil {
		return fmt.Errorf("failed to create new deployment: %v", err)
	}

	// Step 3: Verify the new deployment
	/*fmt.Println("Verifying the new deployment")
	err = verifyDeployment(c, newDeployment.Name, updateForm.Namespace)
	if err != nil {
		return err
	}*/

	// Step 4: Update the service to point to the new deployment
	fmt.Println("Updating the service to point to the new deployment")
	err = updateService(c, service, newDeployment, updateForm)
	if err != nil {
		// Clean up the new deployment if updating the service fails
		deleteErr := DeleteNewDeployment(c, newDeployment.Name, updateForm.Namespace)
		if deleteErr != nil {
			return fmt.Errorf("failed to update service and failed to delete new deployment: %v, %v", err, deleteErr)
		}
		return fmt.Errorf("failed to update service: %v", err)
	}

	// Step 5: Scale down the old deployment
	fmt.Println("Redefining the old deployment")
	err = RedefineOldVersion(c, updateForm.Namespace, deployment.ObjectMeta.Name)
	if err != nil {
		return fmt.Errorf("failed to redefine the old deployment : %v", err)
	}

	return nil

}

// GetDeploymentStatus retrieves the status of the specified deployment
func GetDeploymentStatus(c *gin.Context, deploymentName, namespace string) (*DeploymentStatus, error) {
	clientset := utils.GetClientSet(c)
	deploymentsClient = clientset.AppsV1().Deployments(namespace)

	// Retrieve the deployment
	deployment, err := deploymentsClient.Get(c, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	// Retrieve the pods associated with the deployment
	podsClient := clientset.CoreV1().Pods(namespace)
	podList, err := podsClient.List(c, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	// Collect the status of the pods
	var replicaFailures []string
	for _, pod := range podList.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.State.Waiting != nil {
				replicaFailures = append(replicaFailures, fmt.Sprintf("Pod %s: %s", pod.Name, containerStatus.State.Waiting.Message))
			}
			if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
				replicaFailures = append(replicaFailures, fmt.Sprintf("Pod %s: %s", pod.Name, containerStatus.State.Terminated.Message))
			}
		}
	}

	// Create a DeploymentStatus structure
	status := &DeploymentStatus{
		ReadyReplicas:       deployment.Status.ReadyReplicas,
		AvailableReplicas:   deployment.Status.AvailableReplicas,
		UnavailableReplicas: deployment.Status.UnavailableReplicas,
		UpdatedReplicas:     deployment.Status.UpdatedReplicas,
		ReplicaFailures:     replicaFailures,
		Message:             fmt.Sprintf("Deployment %s in namespace %s has %d ready replicas and %d available replicas", deploymentName, namespace, deployment.Status.ReadyReplicas, deployment.Status.AvailableReplicas),
	}

	return status, nil
}

func verifyDeployment(c *gin.Context, deploymentName, namespace string) error {
	clientset := utils.GetClientSet(c)
	deploymentsClient = clientset.AppsV1().Deployments(namespace)

	//Loop that tries for a maximum of 180 iterations
	for i := 0; i < 180; i++ {
		deployment, err := deploymentsClient.Get(c, deploymentName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		//Checks whether the number of replicas available corresponds to the number of replicas required.
		if deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
			// If the number of available replicas is equal to the number of replicas required, deployment is ready and the function returns nil (no error).
			return nil
		}
		//Wait 1 second before checking again.
		time.Sleep(1 * time.Second)
	}
	//If after 180 iterations the deployment is still not ready, return an error indicating that the deployment is not ready in time.
	return errors.New("new deployment not ready in time")
}

func getServiceByDeployment(c *gin.Context, deployment *v1.Deployment, namespace string) (*apicorev1.Service, error) {
	// Retrieve all services in the specified namespace
	clientset := utils.GetClientSet(c)
	serviceList, err := clientset.CoreV1().Services(namespace).List(c, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Get all labels of the deployment
	deploymentLabels := deployment.Spec.Template.Labels

	// Browse all services to find the one with all selectors matching the deployment labels
	for _, s := range serviceList.Items {
		if len(s.Spec.Selector) == 0 {
			continue
		}
		matches := true
		for key, value := range s.Spec.Selector {
			if deploymentLabels[key] != value {
				//log.Println("Y'a pas de matching.")
				matches = false
				break
			}
		}
		if matches {
			log.Printf("%v", s)
			return &s, nil
		}
	}

	return nil, errors.New("associated service not found")
}

func createNewDeployment(c *gin.Context, deployment *v1.Deployment, updateForm UpdateForm) (*v1.Deployment, error) {
	newDeployment := deployment.DeepCopy()
	newDeployment.ObjectMeta.ResourceVersion = ""
	newDeployment.ObjectMeta.Name = updateForm.Name + "-green"
	newDeployment.Spec.Template.Spec.Containers[0].Image = updateForm.Image
	newDeployment.Spec.Replicas = &updateForm.Replicas
	//newDeployment.Spec.Strategy.Type = updateForm.Strategy

	labels := newDeployment.Spec.Template.Labels
	labels["version"] = "green"
	newDeployment.Spec.Template.Labels = labels

	clientset := utils.GetClientSet(c)
	deploymentsClient = clientset.AppsV1().Deployments(updateForm.Namespace)
	createdDeployment, err := deploymentsClient.Create(c, newDeployment, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return createdDeployment, nil
}

func updateService(c *gin.Context, service *apicorev1.Service, newDeployment *v1.Deployment, updateForm UpdateForm) error {
	clientset := utils.GetClientSet(c)
	servicesClient = clientset.CoreV1().Services(updateForm.Namespace)
	selector := service.Spec.Selector
	selector["version"] = "green"
	service.Spec.Selector = selector

	_, err := servicesClient.Update(c, service, metav1.UpdateOptions{})

	/*if err != nil {
		return err
	}*/
	return err
}

func DeleteNewDeployment(c *gin.Context, newDeploymentName string, namespace string) error {
	clientset := utils.GetClientSet(c)
	deploymentsClient = clientset.AppsV1().Deployments(namespace)
	err := deploymentsClient.Delete(c, newDeploymentName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return errors.New("Failed to delete the deployment")
}

func RedefineOldVersion(c *gin.Context, namespace string, deploymentName string) error {
	clientset := utils.GetClientSet(c)
	deploymentsClient := clientset.AppsV1().Deployments(namespace)

	// Step 1: Retrieve the existing deployment
	deployment, err := deploymentsClient.Get(c, deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %v", err)
	}

	// Step 2: Add the new label to the deployment's pod template labels
	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = make(map[string]string)
	}
	deployment.Spec.Template.Labels["version"] = "blue"

	// Step 3: Update the deployment
	_, err = deploymentsClient.Update(c, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment: %v", err)
	}

	return nil
}

func rollbackToPreviousVersion(c *gin.Context, deployment *v1.Deployment, namespace string) error {
	// Get the Kubernetes clientset
	clientset := utils.GetClientSet(c)

	// Retrieve the service associated with the deployment
	service, err := getServiceByDeployment(c, deployment, namespace)
	if err != nil {
		return fmt.Errorf("failed to get associated service: %v", err)
	}

	// Determine the current and previous version labels
	currentVersion := service.Spec.Selector["version"]
	var previousVersion string
	if currentVersion == "green" {
		previousVersion = "blue"
	} else if currentVersion == "blue" {
		previousVersion = "green"
	} else {
		return errors.New("unknown version label")
	}

	// Update the service selector to point to the previous version
	service.Spec.Selector["version"] = previousVersion

	// Update the service in Kubernetes
	_, err = clientset.CoreV1().Services(namespace).Update(c, service, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update service: %v", err)
	}

	// Optionally, verify that the previous version deployment is ready before returning
	/*previousDeploymentName := fmt.Sprintf("%s-%s", deployment.Name, previousVersion)
	if err := verifyDeployment(c, previousDeploymentName, namespace); err != nil {
		return fmt.Errorf("previous version deployment is not ready: %v", err)
	}*/

	return nil
}
