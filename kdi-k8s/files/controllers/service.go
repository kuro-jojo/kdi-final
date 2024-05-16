package controllers

import (
	"log"
	"net/http"

	"github.com/kuro-jojo/kdi-k8s/files"
	"github.com/kuro-jojo/kdi-k8s/files/objecthandlers"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"

	"github.com/gin-gonic/gin"
)

// CreateService handles the creation request of a kubernetes service from a file
func CreateService(c *gin.Context) {
	namespace, exist := c.GetPostForm("namespace")

	file, err := c.FormFile(files.NameForServiceFileForm)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No file found"})
		return
	}

	objects, co, m := files.ProcessUploadedFile(c, file)
	if co != 0 {
		c.JSON(co, gin.H{"message": m})
		return
	}
	// only handle the Service object for now
	wasServiceCreated := false
	clientset := utils.GetClientSet(c)

	for _, obj := range objects {
		obj, ok := obj.(*models.Service)
		if !ok {
			continue
		}
		if obj != nil {
			if exist && namespace != "" {
				obj.Service.Namespace = namespace
			}
			obj.Clientset = clientset
			co, m = objecthandlers.HandleKubeObjectCreation(obj, c)
			if co != http.StatusOK {
				c.JSON(co, gin.H{"message": m})
				return
			}
			wasServiceCreated = true
		}
	}
	if !wasServiceCreated {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No service object found in the file"})
	} else {
		c.JSON(co, gin.H{"message": m})
	}
	log.Println("Service created successfully")
}
