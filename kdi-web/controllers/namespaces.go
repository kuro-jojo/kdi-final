package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetNamespacesFromCluster gets all namespaces from the cluster by calling the kubernetes API
func GetNamespacesFromCluster(c *gin.Context) {
	log.Println("Getting all namespaces...")

	user, driver := GetUserFromContext(c)

	c_id := c.Param("id")
	id, err := primitive.ObjectIDFromHex(c_id)
	if err != nil {
		log.Printf("Error converting id %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid id"})
		return
	}
	cluster := models.Cluster{
		ID: id,
	}

	ok := UserHasRightOnCluster(c, driver, cluster, user, []string{models.ListNamespacesRole})
	if !ok {
		return
	}
	err = cluster.Get(driver)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		}
		return
	}

	resp, body, ok := MakeRequestToKubernetesAPI(c, cluster, "GET", "/resources/namespaces", c.Request.Body)
	if !ok {
		return
	}

	response := make(map[string][]string)
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Error unmarshalling response %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error unmarshalling response"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error making request : %v", response["message"])
		c.JSON(resp.StatusCode, response)
		return
	}

	c.JSON(http.StatusOK, response)
}
