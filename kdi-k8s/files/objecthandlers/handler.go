package objecthandlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/models"
	"github.com/kuro-jojo/kdi-k8s/utils"
)

// HandleKubeObjectCreation handles the creation of a kubernetes object
func HandleKubeObjectCreation(obj models.KubeObject, c *gin.Context) (int, string) {
	log.Printf("Creating %T: %s\n", obj, obj.GetName())

	if obj.GetNamespace() == "" {
		log.Printf("Namespace not provided for object %s. Setting it to \"default\"\n", obj.GetName())
		obj.SetNamespace("default")
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
			// TODO : handle other errors
			return http.StatusInternalServerError, "An unexpected error occurred"
		}
	} else {
		log.Printf("Object already exists: %s\n", obj.GetName())
		return http.StatusConflict, fmt.Sprintf("%s already exists in namespace %s", obj.GetName(), obj.GetNamespace())
	}

	log.Printf("Creating object: %s in namespace %s\n", obj.GetName(), obj.GetNamespace())
	for {
		err := obj.Create(context.TODO(), obj, metav1.CreateOptions{})
		if err != nil {
			// Create namespace if it doesn't exist
			if yes := utils.IsNotFoundError(err.Error()); yes {
				ns := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: obj.GetNamespace(),
					},
				}
				_, err := utils.GetClientSet(c).CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
				if err != nil {
					log.Printf("Error on creating namespace %s: %v\n", obj.GetNamespace(), err)
					return http.StatusInternalServerError, fmt.Sprintf("Error on creating namespace %s", obj.GetNamespace())
				}
				log.Printf("Namespace created: %s\n", obj.GetNamespace())
				continue
			}
			log.Printf("Error on creating object %s in namespace %s: %v\n", obj.GetName(), obj.GetNamespace(), err)
			e := fmt.Sprintf("Error on creating object %s in namespace %s\n", obj.GetName(), obj.GetNamespace())
			return http.StatusInternalServerError, e
		}
		log.Printf("Object created: %s\n", obj.GetName())
		return http.StatusCreated, fmt.Sprintf("%s created successfully in namespace %s", obj.GetName(), obj.GetNamespace())
	}
}
