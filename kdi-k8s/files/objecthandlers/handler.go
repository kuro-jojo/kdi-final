package objecthandlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
)

func HandleKubeObjectCreation(obj models.KubeObject, c *gin.Context) (int, string) {
	log.Printf("\t\tCreating %T: %s\n", obj, obj.GetName())

	if obj.GetNamespace() == "" {
		return http.StatusBadRequest, "Namespace is required for creation"
	}

	err := obj.Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
	if err != nil {
		if yes, e := utils.IsUnauthorizedError(err.Error(), obj.GetName()); yes {
			log.Printf("%s in namespace %s. Namespace doesn't exist or Forbidden\n", e.Error(), obj.GetNamespace())
			return http.StatusUnauthorized, fmt.Sprintf("%s in namespace %s. Namespace doesn't exist or Forbidden\n", e.Error(), obj.GetNamespace())
		} else if yes, _ := utils.IsForbiddenError(err.Error(), obj.GetName()); yes {
			log.Printf("Cannot access %s in the namespace %s\n", obj.GetName(), obj.GetNamespace())
			return http.StatusForbidden, fmt.Sprintf("Cannot access %s in the namespace %s", obj.GetName(), obj.GetNamespace())
		} else if !utils.IsNotFoundError(err.Error()) {
			log.Printf("Error on getting object: %v.\n", err)
			return http.StatusInternalServerError, "An unexpected error occurred"
		}
	} else {
		log.Printf("Object already exists: %s\n", obj.GetName())
		return http.StatusConflict, fmt.Sprintf("%s already exists in namespace %s", obj.GetName(), obj.GetNamespace())
	}
	log.Printf("Creating object: %s\n", obj.GetName())
	newObj, err := obj.Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error on creating object %s in namespace %s: %v\n", obj.GetName(), obj.GetNamespace(), err)
		e := fmt.Sprintf("Error on creating object: %s\n", obj.GetName())
		return http.StatusInternalServerError, e
	}

	log.Printf("Object created: %s\n", newObj.GetName())
	return http.StatusOK, fmt.Sprintf("%s created successfully.", newObj.GetName())
}
