package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/files"
	"github.com/kuro-jojo/kdi-k8s/files/objecthandlers"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TimeToWaitForGettingDeploymentStatus = 1 * time.Second
)

// CreateMultipleRessources handles the creation request of multiple kubernetes resources from a file
func CreateMultipleRessources(c *gin.Context) {
	namespace, exist := c.GetPostForm("namespace")
	if exist {
		log.Printf("Setting default namespace to %s", namespace)
	}
	type Response struct {
		Messages      map[string][]string   `json:"messages"`
		Microservices []models.Microservice `json:"microservices"`
	}
	var response Response
	response.Messages = make(map[string][]string, 0)
	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Error getting the form : %v", err)
		response.Messages["error"] = append(response.Messages["error"], "Error getting the form")
		c.JSON(http.StatusBadRequest, gin.H{"messages": response.Messages})
		return
	}

	uploadedFiles := form.File[files.NameForFilesForm]
	if len(uploadedFiles) == 0 {
		log.Printf("No file uploaded")
		response.Messages["error"] = append(response.Messages["error"], "No file uploaded")
		c.JSON(http.StatusNotFound, gin.H{"messages": response.Messages})
		return
	}
	var httpResps = make(map[int][]string, 0)
	clientset := utils.GetClientSet(c)

	for _, file := range uploadedFiles {
		objects, co, m := files.ProcessUploadedFile(c, file)
		if co != 0 {
			httpResps[co] = append(httpResps[co], m)
			continue
		}

		for _, obj := range objects {
			isDeployment := false
			switch obj := obj.(type) {
			case *models.Deployment:
				if obj.Deployment != nil {
					if exist && namespace != "" {
						obj.Deployment.Namespace = namespace
					}
					obj.Clientset = clientset
					isDeployment = true
				}
			case *models.Service:
				if obj.Service != nil {
					if exist && namespace != "" {
						obj.Service.Namespace = namespace
					}
					obj.Clientset = clientset
				}
			}

			co, m = objecthandlers.HandleKubeObjectCreation(obj, c)
			httpResps[co] = append(httpResps[co], m+" (file : "+file.Filename+")")

			if isDeployment && co == http.StatusCreated {
				o, ok := obj.(*models.Deployment)
				if !ok {
					log.Printf("Error casting object to deployment")
					continue
				}
				// waiting a few seconds to get the deployment status after creation
				time.Sleep(TimeToWaitForGettingDeploymentStatus)
				err = o.Get(context.TODO(), o.GetName(), metav1.GetOptions{})
				if err != nil {
					log.Printf("Error getting deployment %s: %v", o.GetName(), err)
					continue
				}
				conditions := make([]models.Conditions, 0)
				for _, c := range o.Deployment.Status.Conditions {
					conditions = append(conditions, models.Conditions{
						Type:    string(c.Type),
						Message: c.Message,
						Reason:  c.Reason,
					})
				}

				containers := make([]models.Container, 0)
				for _, c := range o.Deployment.Spec.Template.Spec.Containers {
					containers = append(containers, models.Container{
						Name:  c.Name,
						Image: c.Image,
					})
					if len(c.Ports) > 0 {
						containers[len(containers)-1].Port = c.Ports[0].ContainerPort
					}
				}

				response.Microservices = append(response.Microservices, models.Microservice{
					Name:       o.GetName(),
					Namespace:  o.GetNamespace(),
					Replicas:   *o.Deployment.Spec.Replicas,
					Labels:     o.Deployment.Spec.Template.Labels,
					Selectors:  o.Deployment.Spec.Selector.MatchLabels,
					Strategy:   string(o.Deployment.Spec.Strategy.Type),
					Conditions: conditions,
					Containers: containers,
				})

				log.Printf("Microservice %s created successfully", o.GetName())
			}
		}
	}

	status := utils.GetMostSeenCode(httpResps)
	if status == -1 {
		log.Printf("No valid kubernetes object found in all the files")
		response.Messages["error"] = append(response.Messages["error"], "No valid kubernetes object found in  all the files")
		c.JSON(http.StatusBadRequest, gin.H{"messages": response.Messages})
		return
	}
	if code := httpResps[http.StatusOK]; len(code) > 0 {
		status = http.StatusOK
	}

	for co, v := range httpResps {
		if co == http.StatusCreated {
			response.Messages["success"] = append(response.Messages["success"], v...)
		} else {
			response.Messages["error"] = append(response.Messages["error"], v...)
		}
	}
	c.JSON(status, gin.H{"messages": response.Messages, "microservices": response.Microservices, "size": len(response.Microservices)})
}
