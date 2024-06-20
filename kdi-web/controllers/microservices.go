package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	NameForFilesForm = "files"
	// TODO: Change this for production
)

type K8sApiHttpResponse struct {
	Messages      map[string][]string   `json:"messages"`
	Message       string                `json:"message"`
	Microservices []models.Microservice `json:"microservices"`
}

type MicroserviceUpdateForm struct {
	Name      string `json:"name" `
	Namespace string `json:"namespace" `
	Strategy  string `json:"strategy" binding:"required"`
	Container string `json:"container"`
	Image     string `json:"image" binding:"required"`
	Replicas  int32  `json:"replicas" binding:"required"`
	// Add more fields here depending on the strategy

	// For rolling update strategy
	MaxUnavailable string `json:"maxUnavailable"` // The maximum number of pods that can be unavailable during the update process
	MaxSurge       string `json:"maxSurge"`       // The maximum number of pods that can be scheduled above the desired number of pods
}

func CreateMicroserviceWithYaml(c *gin.Context) {
	kubernetesApiUrl := os.Getenv("KDI_K8S_API_ENDPOINT")

	// retrieve the cluster from the environment
	user, driver := GetUserFromContext(c)
	eId := c.Param("e_id")
	id, err := primitive.ObjectIDFromHex(eId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid environment ID"})
		return
	}

	// 1. Get the environment
	environment := models.Environment{
		ID: id,
	}
	err = environment.Get(driver)
	if err != nil {
		log.Printf("Error getting environment %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environment"})
		return
	}

	// 2. Get the project and check if the user has access to it
	p_id, err := primitive.ObjectIDFromHex(environment.ProjectID)
	if err != nil {
		log.Printf("Invalid project ID : %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	project := models.Project{
		ID: p_id,
	}

	err = project.Get(driver)
	if err != nil {
		log.Printf("Error getting project %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting project"})
		return
	}
	if project.TeamspaceID == "" && project.CreatorID != user.ID.Hex() {
		log.Printf("Unauthorized: Not enough privilege to make deployments")
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	// Check if the user has enough privilege in the teamspace to make deployments
	if project.TeamspaceID != "" {
		t_id, err := primitive.ObjectIDFromHex(project.TeamspaceID)
		if err != nil {
			log.Printf("Invalid teamspace ID : %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
			return
		}
		teamspace := models.Teamspace{
			ID: t_id,
		}
		err = teamspace.Get(driver)
		if err != nil {
			log.Printf("Error getting teamspace %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting teamspace"})
			return
		}
		ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.CreateDeploymentRole}, teamspace, user)
		if !ok {
			log.Println(message)
			c.JSON(code, gin.H{"message": message})
			return
		}
	}

	// 3. Get the cluster and make a request to the kubernetes api to create the microservice
	c_id, err := primitive.ObjectIDFromHex(environment.ClusterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster := models.Cluster{
		ID: c_id,
	}

	err = cluster.Get(driver)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		return
	}

	if slices.Contains(cluster.Teamspaces, project.TeamspaceID) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// Make a request to the kubernetes api
	req, err := http.NewRequest("POST", kubernetesApiUrl+"/resources/with-yaml", c.Request.Body)
	if err != nil {
		log.Printf("Error creating request %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error making deployments"})
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
		log.Printf("Error making deployments %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error making deployments"})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body %v", err)
		c.JSON(resp.StatusCode, gin.H{"message": "Error making deployments"})
		return
	}

	var r K8sApiHttpResponse
	r.Messages = make(map[string][]string)
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Printf("Error unmarshalling response body %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error making deployments"})
		return
	}

	// 4. Save the microservice in the database
	for _, m := range r.Microservices {
		// 4.1 Create the namespace in the database if it does not exist
		ns := models.Namespace{
			// The namespace in the microservice object at this point is just the name of the namespace and not the ID
			Name:      m.NamespaceID,
			ClusterID: cluster.ID.Hex(),
		}

		// Check if the namespace exists
		err = ns.GetByName(driver)
		if err != nil {
			if utils.OnNotFoundError(err, "Namespace") != nil || utils.OnNoDocumentsError(err, "Namespace") != nil {
				log.Printf("Namespace %s does not exist. Creating it", ns.Name)
				err = ns.Create(driver)
				if err != nil {
					// There won't be a duplicate key error here because we already checked if the namespace exists
					log.Printf("Error creating namespace %s : %v", ns.Name, err)
					r.Messages["error"] = append(r.Messages["error"], "Error creating namespace "+ns.Name)
					continue
				}
			} else {
				log.Printf("Error getting namespace %s : %v", ns.Name, err)
				r.Messages["error"] = append(r.Messages["error"], "Error getting namespace "+ns.Name)
				continue
			}
		}

		// 4.2 Save the microservice
		m.EnvironmentID = eId
		m.CreatorID = user.ID.Hex()
		m.NamespaceID = ns.ID.Hex()
		m.DeployedAt = time.Now()

		err = m.Create(driver)
		if err != nil {
			log.Printf("Error creating microservice %v", err)
			if er := utils.OnDuplicateKeyError(err, "Microservice"); er != nil {
				r.Messages["info"] = append(r.Messages["info"], "Microservice "+m.Name+" already saved")
			} else {
				r.Messages["error"] = append(r.Messages["error"], "Error saving microservice "+m.Name)
			}
			continue
		}
		log.Printf("Microservice %s saved successfully", m.Name)
		r.Messages["success"] = append(r.Messages["success"], "Microservice "+m.Name+" saved successfully")
	}

	if r.Message != "" {
		r.Messages["error"] = append(r.Messages["error"], r.Message)
	}
	c.JSON(resp.StatusCode, gin.H{"messages": r.Messages, "microservices": r.Microservices})
}

func GetMicroservices(c *gin.Context) {
	user, driver := GetUserFromContext(c)

	m := models.Microservice{
		CreatorID: user.ID.Hex(),
	}
	microservices, err := m.GetAllByCreator(driver)
	if err != nil {
		log.Printf("Error getting microservices %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting microservices"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"microservices": microservices, "size": len(microservices)})
}

func GetMicroserviceByEnvironment(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("m_id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid microservice ID"})
		return
	}
	// Get the environment ID
	eId := c.Param("e_id")

	microservice := models.Microservice{
		ID:            objectID,
		CreatorID:     user.ID.Hex(),
		EnvironmentID: eId,
	}

	err = microservice.Get(driver)

	if err != nil {
		log.Printf("Error getting microservice %v", err)
		if er := utils.OnDuplicateKeyError(err, "Microservice"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"microservice": microservice})
}

func GetMicroservicesByEnvironment(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	eId := c.Param("e_id")

	m := models.Microservice{
		CreatorID:     user.ID.Hex(),
		EnvironmentID: eId,
	}

	microservices, err := m.GetAllByEnvironment(driver)
	if err != nil {
		log.Printf("Error getting microservices %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting microservices"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"microservices": microservices, "size": len(microservices)})
}

func UpdateMicroservice(c *gin.Context) {
	kubernetesApiUrl := os.Getenv("KDI_K8S_API_ENDPOINT")

	// Prepare the update form
	var updateForm MicroserviceUpdateForm
	if err := c.ShouldBindJSON(&updateForm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid update form"})
		return
	}

	_, driver := GetUserFromContext(c)

	id := c.Param("m_id")
	e_id := c.Param("e_id")
	m_id, _ := primitive.ObjectIDFromHex(id)
	env_id, err := primitive.ObjectIDFromHex(e_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid microservice or environment ID"})
		return
	}

	// 1. Get the microservice
	microservice := models.Microservice{
		ID: m_id,
	}

	err = microservice.Get(driver)
	if err != nil {
		log.Printf("Error getting microservice %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting microservice"})
		return
	}

	// 2. Get the environment
	environment := models.Environment{
		ID: env_id,
	}
	err = environment.Get(driver)
	if err != nil {
		log.Printf("Error getting environment %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environment"})
		return
	}

	// 2. Get the namespace
	namespace_id, err := primitive.ObjectIDFromHex(microservice.NamespaceID)
	if err != nil {
		log.Printf("Invalid namespace ID : %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid namespace ID"})
		return
	}

	n := models.Namespace{
		ID: namespace_id,
	}

	err = n.Get(driver)
	if err != nil {
		log.Printf("Error getting namespace %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting namespace"})
		return
	}

	// 3. Get the cluster and make a request to the kubernetes api to create the microservice
	c_id, err := primitive.ObjectIDFromHex(environment.ClusterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster := models.Cluster{
		ID: c_id,
	}

	err = cluster.Get(driver)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		return
	}

	namespace := n.Name
	deploymentName := microservice.Name

	// Serialize the updateForm to JSON for the request body
	updateFormJSON, err := json.Marshal(updateForm)
	if err != nil {
		log.Printf("Error marshalling update form to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error preparing update data"})
		return
	}

	// Log the serialized JSON for debugging purposes
	log.Printf("UpdateForm JSON: %s", updateFormJSON)

	// Make a request to the kubernetes api
	req, err := http.NewRequest("PATCH", kubernetesApiUrl+"/resources/namespaces/"+namespace+"/deployments/"+deploymentName, bytes.NewBuffer(updateFormJSON))
	if err != nil {
		log.Printf("Error creating request %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating deployment"})
		return
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Authorization", cluster.Token)

	// Create a client
	client := &http.Client{
		Timeout: 120 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making deployments %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error making deployments"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error from Kubernetes API: %v", string(bodyBytes))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error from Kubernetes API", "details": string(bodyBytes)})
		return
	}

	if updateForm.Strategy == "blue-green" {

		// Only update the microservice information if the Kubernetes API call was successful
		microservice.Replicas = updateForm.Replicas
		microservice.DeployedAt = time.Now()
		microservice.Labels["version"] = "green"
		microservice.Name = microservice.Name + "-green"
		for i := range microservice.Containers {
			microservice.Containers[i].Image = updateForm.Image
		}

	} else {
		// Only update the microservice information if the Kubernetes API call was successful
		microservice.Replicas = updateForm.Replicas
		microservice.Strategy = updateForm.Strategy
		microservice.DeployedAt = time.Now()
		for i := range microservice.Containers {
			microservice.Containers[i].Image = updateForm.Image
		}
	}

	err = microservice.Update(driver)
	if err != nil {
		log.Printf("Error updating microservice %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error updating microservice"})
		return
	}

	log.Printf("Microservice %s updated successfully", microservice.Name)
	c.JSON(resp.StatusCode, gin.H{"message": "Microservice updated successfully", "microservice": microservice})
}
