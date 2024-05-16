package controllers

import (
	"net/http"
	"strings"

	"github.com/kuro-jojo/kdi-k8s/files"
	"github.com/kuro-jojo/kdi-k8s/files/objecthandlers"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"

	"github.com/gin-gonic/gin"
)

// CreateMultipleRessources handles the creation request of multiple kubernetes resources from a file
func CreateMultipleRessources(c *gin.Context) {
	namespace, exist := c.GetPostForm("namespace")

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	uploadedFiles := form.File[files.NameForFilesForm]
	if len(uploadedFiles) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No file uploaded"})
		return
	}

	var httpResps = make(map[int][]string, 0)
	clientset := utils.GetClientSet(c)

	for _, file := range uploadedFiles {
		objects, co, m := files.ProcessUploadedFile(c, file)
		if co != 0 {
			c.String(co, m)
			httpResps[co] = append(httpResps[co], m)
			continue
		}

		for _, obj := range objects {
			switch obj := obj.(type) {
			case *models.Deployment:
				if obj.Deployment != nil {
					if exist && namespace != "" {
						obj.Deployment.Namespace = namespace
					}
					obj.Clientset = clientset
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
			httpResps[co] = append(httpResps[co], m)
		}
	}

	status := utils.GetMostSeenCode(httpResps)
	if status == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "No valid kubernetes object found in the file"})
		return
	}
	if code := httpResps[http.StatusOK]; len(code) > 0 {
		status = http.StatusOK
	}

	var r []string
	for _, v := range httpResps {
		r = append(r, strings.Join(v, "   "))
	}
	c.JSON(status, gin.H{"message": strings.Join(r, "   ")})
}
