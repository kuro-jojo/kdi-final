package namespaces

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNamespaces(c *gin.Context) {
	log.Println("Getting all namespaces from cluster...")

	clientset := utils.GetClientSet(c)
	namespaces, err := clientset.CoreV1().Namespaces().List(c, metav1.ListOptions{
		Limit: 20,
	})

	if err != nil {
		log.Println("Error getting namespaces: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	log.Println("Namespaces found: ", len(namespaces.Items))
	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces.Items,
	})
}
