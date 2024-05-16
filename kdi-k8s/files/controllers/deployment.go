package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/files"
	"github.com/kuro-jojo/kdi-k8s/files/objecthandlers"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"
)

// CreateDeployment handles the creation request of a kubernetes deployment from a file
func CreateDeployment(c *gin.Context) {
	namespace, exist := c.GetPostForm("namespace")

	file, err := c.FormFile(files.NameForDeploymentFileForm)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No file found"})
		return
	}
	objects, co, m := files.ProcessUploadedFile(c, file)
	if co != 0 {
		c.JSON(co, gin.H{"message": m})
		return
	}

	wasDeploymentCreated := false
	clientset := utils.GetClientSet(c)

	for _, obj := range objects {
		obj, ok := obj.(*models.Deployment)
		if !ok {
			continue
		}
		if obj != nil {
			if exist && namespace != "" {
				obj.Deployment.Namespace = namespace
			}
			obj.Clientset = clientset
			co, m = objecthandlers.HandleKubeObjectCreation(obj, c)
			if co != http.StatusOK {
				c.JSON(co, gin.H{"message": m})
				return
			}
			wasDeploymentCreated = true
		}
	}

	if !wasDeploymentCreated {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No deployment object found in the file"})
	} else {
		c.JSON(co, gin.H{"message": m})
	}
	log.Println("Deployment created successfully")
}
