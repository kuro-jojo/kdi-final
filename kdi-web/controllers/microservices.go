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
	log.Println("Creating microservice with yaml files ...")
	kubernetesApiUrl := os.Getenv("KDI_K8S_API_ENDPOINT")
	var r K8sApiHttpResponse
	r.Messages = make(map[string][]string)
	// retrieve the cluster from the environment
	user, driver := GetUserFromContext(c)

	eId := c.Param("e_id")
	id, err := primitive.ObjectIDFromHex(eId)
	if err != nil {
		r.Messages["error"] = append(r.Messages["error"], "Invalid environment ID")
		c.JSON(http.StatusBadRequest, gin.H{"messages": r.Messages})
		return
	}

	// 1. Get the environment
	environment := models.Environment{
		ID: id,
	}
	err = environment.Get(driver)
	if err != nil {
		log.Printf("Error getting environment %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error getting environment")
		c.JSON(http.StatusInternalServerError, gin.H{"messages": r.Messages})
		return
	}

	// 2. Get the project and check if the user has access to it
	p_id, err := primitive.ObjectIDFromHex(environment.ProjectID)
	if err != nil {
		log.Printf("Invalid project ID : %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Invalid project ID")
		c.JSON(http.StatusBadRequest, gin.H{"messages": r.Messages})
		return
	}

	project := models.Project{
		ID: p_id,
	}

	log.Println("Getting project ...")
	err = project.Get(driver)
	if err != nil {
		log.Printf("Error getting project %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error getting project")
		c.JSON(http.StatusBadRequest, gin.H{"messages": r.Messages})
		return
	}
	// Check if the user has enough privilege in the project to make deployments
	if project.CreatorID != user.ID.Hex() {
		if project.TeamspaceID == "" && project.CreatorID != user.ID.Hex() {
			log.Printf("Unauthorized: Cannot make deployments to a project you do not own")
			r.Messages["error"] = append(r.Messages["error"], "Unauthorized: Cannot make deployments to a project you do not own")
			c.JSON(http.StatusUnauthorized, gin.H{"messages": r.Messages})
			return
		}
		// Check if the user has enough privilege in the teamspace to make deployments
		if project.TeamspaceID != "" {
			t_id, err := primitive.ObjectIDFromHex(project.TeamspaceID)
			if err != nil {
				log.Printf("Invalid teamspace ID : %v", err)
				r.Messages["error"] = append(r.Messages["error"], "Invalid teamspace ID")
				c.JSON(http.StatusBadRequest, gin.H{"messages": r.Messages})
				return
			}
			teamspace := models.Teamspace{
				ID: t_id,
			}
			err = teamspace.Get(driver)
			if err != nil {
				log.Printf("Error getting teamspace %v", err)
				r.Messages["error"] = append(r.Messages["error"], "Error getting teamspace")
				c.JSON(http.StatusInternalServerError, gin.H{"messages": r.Messages})
				return
			}
			ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.CreateDeploymentRole}, teamspace, user)
			if !ok {
				log.Println(message)
				r.Messages["error"] = append(r.Messages["error"], message)
				c.JSON(code, gin.H{"messages": r.Messages})
				return
			}
		}
	}
	// 3. Get the cluster and make a request to the kubernetes api to create the microservice
	c_id, err := primitive.ObjectIDFromHex(environment.ClusterID)
	if err != nil {
		r.Messages["error"] = append(r.Messages["error"], "Invalid cluster ID")
		c.JSON(http.StatusBadRequest, gin.H{"messages": r.Messages})
		return
	}

	cluster := models.Cluster{
		ID: c_id,
	}

	err = cluster.Get(driver)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error getting cluster")
		c.JSON(http.StatusInternalServerError, gin.H{"messages": r.Messages})
		return
	}

	if slices.Contains(cluster.Teamspaces, project.TeamspaceID) {
		r.Messages["error"] = append(r.Messages["error"], "Unauthorized")
		c.JSON(http.StatusUnauthorized, gin.H{"messages": r.Messages})
		return
	}

	// Make a request to the kubernetes api
	req, err := http.NewRequest("POST", kubernetesApiUrl+"/resources/with-yaml", c.Request.Body)
	if err != nil {
		log.Printf("Error creating request %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error making deployments on the cluster")
		c.JSON(http.StatusInternalServerError, gin.H{"messages": r.Messages})
		return
	}
	req.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	req.Header.Set("Authorization", cluster.Token)
	req.Header.Set("cluster-type", cluster.Type)

	// Create a client
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making deployments %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error making deployments on the cluster")
		c.JSON(http.StatusInternalServerError, gin.H{"messages": r.Messages})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error reading response body")
		c.JSON(resp.StatusCode, gin.H{"messages": r.Messages})
		return
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Printf("Error unmarshalling response body %v", err)
		r.Messages["error"] = append(r.Messages["error"], "Error unmarshalling response body")
		c.JSON(resp.StatusCode, gin.H{"messages": r.Messages, "microservices": r.Microservices})
		return
	}

	// TODO : Save the microservices in the database even if it already exists in the cluster
	// 4. Save the microservice in the database
	for _, m := range r.Microservices {

		// 4.1 Save the microservice
		m.EnvironmentID = eId
		m.CreatorID = user.ID.Hex()
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
	log.Println("Microservices created successfully")
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

	// Serialize the updateForm to JSON for the request body
	updateFormJSON, err := json.Marshal(updateForm)
	if err != nil {
		log.Printf("Error marshalling update form to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error preparing update data"})
		return
	}

	// Make a request to the kubernetes api
	resp, body, ok := MakeRequestToKubernetesAPI(c, cluster, "PATCH", "/resources/namespaces/"+microservice.Namespace+"/deployments/"+microservice.Name, bytes.NewBuffer(updateFormJSON))
	if !ok {
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error from Kubernetes API: %v", string(body))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error from Kubernetes API", "details": string(body)})
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
