package deployments

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeploymentForm struct {
	DeploymentName string
	Namespace      string
}

func GetDeploymentInNamespace(c *gin.Context) {
	log.Println("Getting deployment...")

	deploymentForm := DeploymentForm{
		DeploymentName: c.Param("deployment"),
		Namespace:      c.Param("namespace"),
	}
	if !isDeploymentFormValid(deploymentForm) {
		c.JSON(http.StatusBadRequest, gin.H{"message": " Please provide deployment_name and namespace"})
	}

	clientset := utils.GetClientSet(c)
	deployment, err := clientset.AppsV1().Deployments(deploymentForm.Namespace).Get(context.TODO(), deploymentForm.DeploymentName, metav1.GetOptions{})
	if err != nil && utils.IsNotFoundError(err.Error()) {
		log.Printf("Error getting deployment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("deployment %s not found in namespace %s", deploymentForm.DeploymentName, deploymentForm.Namespace)})
	}

	log.Printf("Deployment %s found in namespace %s", deploymentForm.DeploymentName, deploymentForm.Namespace)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("deployment %s found in namespace %s", deploymentForm.DeploymentName, deploymentForm.Namespace), "deployment": deployment})
}

func isDeploymentFormValid(form DeploymentForm) bool {
	return form.DeploymentName != "" && form.Namespace != ""
}
