package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetNamespace(c *gin.Context) {
	_, driver := GetUserFromContext(c)
	id := c.Param("n_id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	namespace := models.Namespace{
		ID: objectID,
		//CreatorID: user.ID.Hex(),
	}

	err := namespace.Get(driver)

	if err != nil {
		log.Printf("Error getting namespace %v", err)
		if er := utils.OnDuplicateKeyError(err, "Namespace"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": message})
	c.JSON(http.StatusOK, gin.H{"namespace": namespace.Name})
}

// GetNamespacesFromCluster gets all namespaces from the cluster by calling the kubernetes API
func GetNamespacesFromCluster(c *gin.Context) {
	log.Println("Getting all namespaces...")
	
	kubernetesApiUrl := os.Getenv("KDI_K8S_API_ENDPOINT")

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

	req, err := http.NewRequest("GET", kubernetesApiUrl+"/resources/namespaces", c.Request.Body)
	if err != nil {
		log.Printf("Error creating request %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error creating request"})
		return
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Authorization", cluster.Token)

	// Create a client
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error making request"})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error reading response body"})
		return
	}

	response := make(map[string]string)
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
