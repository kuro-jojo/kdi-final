package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClusterForm struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Address     string   `json:"address"`
	Port        string   `json:"port"`
	Token       string   `json:"token"`
	Teamspaces  []string `json:"teamspaces"`
	IsGlobal    bool     `json:"isGlobal"`
	CreatedAt   time.Time
}

func TestConnectionToCluster(c *gin.Context) {
	log.Println("Testing connection to cluster...")
	kubernetesApiUrl := os.Getenv("KDI_K8S_API_ENDPOINT")

	var clusterForm ClusterForm
	if c.BindJSON(&clusterForm) != nil {
		log.Printf("Invalid form : %v", c.BindJSON(&clusterForm).Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if clusterFormIsInValid(clusterForm) {
		log.Println("Invalid form fields")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	cluster := models.Cluster{
		Address: clusterForm.Address,
		Port:    clusterForm.Port,
	}
	exp, err := GetTokenExpirationDate(clusterForm.Token)
	if err != nil {
		log.Printf("Error getting token expiration date: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	cluster.ExpiryDate = exp
	token, err := generateClusterJWT(cluster, clusterForm.Token)
	if err != nil {
		log.Printf("Error generating cluster token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	cluster.Token = token
	// Make a request to the kubernetes api
	req, err := http.NewRequest("GET", kubernetesApiUrl+"/auth", c.Request.Body)
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

	if resp.StatusCode != 200 {
		message := make(map[string]string)
		err = json.Unmarshal(body, &message)
		if err != nil {
			log.Printf("Error unmarshalling response body %v", err)
			c.JSON(resp.StatusCode, gin.H{"message": string(body)})
			return
		}
		log.Printf("Error making request : %v", message["message"])
		c.JSON(resp.StatusCode, message)
		return
	}

	log.Println("Connection to cluster successful")
	c.JSON(http.StatusOK, gin.H{"message": "Connection to cluster successful"})
}

func AddCluster(c *gin.Context) {
	log.Println("Creating cluster...")

	user, driver := GetUserFromContext(c)

	var clusterForm ClusterForm
	if c.BindJSON(&clusterForm) != nil {
		log.Printf("Invalid form : %v", c.BindJSON(&clusterForm).Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if clusterFormIsInValid(clusterForm) {
		log.Println("Invalid form fields")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	cluster, code, message := setupCluster(driver, clusterForm, user)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	cluster.CreatedAt = time.Now()
	err := cluster.Add(driver)
	if err != nil {
		log.Printf("Error creating cluster %v", err)
		if er := utils.OnDuplicateKeyError(err, "Cluster"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	log.Println("Cluster created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Cluster created successfully"})
}

func UpdateCluster(c *gin.Context) {
	log.Println("Editing cluster...")

	user, driver := GetUserFromContext(c)

	var clusterForm ClusterForm
	if c.BindJSON(&clusterForm) != nil {
		log.Printf("Invalid form : %v", c.BindJSON(&clusterForm).Error())
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if clusterFormIsInValid(clusterForm) {
		log.Println("Invalid form fields")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Printf("Invalid cluster ID %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster, code, message := setupCluster(driver, clusterForm, user)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	cluster.ID = id

	ok := UserHasRightOnCluster(c, driver, cluster, user, []string{models.UpdateTeamspaceRole})
	if !ok {
		return
	}

	err = cluster.Update(driver)
	if err != nil {
		log.Printf("Error creating cluster %v", err)
		if er := utils.OnDuplicateKeyError(err, "Cluster"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	log.Println("Cluster edited successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Cluster edited successfully"})
}

func DeleteCluster(c *gin.Context) {
	log.Println("Deleting cluster...")

	user, driver := GetUserFromContext(c)

	clusterID := c.Param("id")
	if clusterID == "" {
		log.Println("Invalid cluster ID")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Printf("Invalid cluster ID %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster := models.Cluster{
		ID: id,
	}

	ok := UserHasRightOnCluster(c, driver, cluster, user, []string{models.DeleteClusterRole})
	if !ok {
		return
	}

	err = cluster.Delete(driver)
	if err != nil {
		log.Printf("Error deleting cluster %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error deleting cluster"})
		}
		return
	}
	log.Println("Cluster deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Cluster deleted successfully"})
}

// GetClustersByCreator returns all clusters of the current user
func GetClustersByCreator(c *gin.Context) {
	log.Println("Listing all clusters of the current user...")
	user, driver := GetUserFromContext(c)

	cluster := models.Cluster{
		CreatorID: user.ID.Hex(),
	}

	clusters, err := cluster.GetAllByCreator(driver)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		}
		return
	}
	log.Println("Clusters retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"clusters": clusters, "size": len(clusters)})
}

// GetClusterByIDAndCreator returns a cluster by its ID only if the creator is the one making the request
func GetClusterByIDAndCreator(c *gin.Context) {
	log.Println("Getting cluster by ID...")
	user, driver := GetUserFromContext(c)

	clusterID := c.Param("id")
	if clusterID == "" {
		log.Println("Invalid cluster ID")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}
	id, err := primitive.ObjectIDFromHex(clusterID)
	if err != nil {
		log.Printf("Invalid cluster ID %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster := models.Cluster{
		ID: id,
	}

	err = cluster.GetByCreator(driver, user.ID)
	if err != nil {
		log.Printf("Error getting cluster %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		}
		return
	}

	queryParam := c.Query("token")
	if queryParam != "" && queryParam == "false" {
		// Retreive the cluster token from the JWT token
		token, err := GetClusterTokenFromJWT(cluster.Token)
		if err != nil {
			log.Printf("Error getting cluster token from JWT %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster token from JWT"})
			return
		}
		cluster.Token = token
	}

	log.Println("Cluster retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"cluster": cluster})
}

func GetClusterName(c *gin.Context) {

	_, driver := GetUserFromContext(c)
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	cluster := models.Cluster{
		ID: objectID,
		//CreatorID: user.ID.Hex(),
	}

	err := cluster.Get(driver)

	if err != nil {
		log.Printf("Error getting cluster %v", err)
		if er := utils.OnDuplicateKeyError(err, "Cluster"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": message})
	c.JSON(http.StatusOK, gin.H{"cluster": cluster.Name})

}

// GetClustersByTeamspace returns all clusters of the teamspace the user is in
func GetClustersByTeamspace(c *gin.Context) {
	log.Println("Getting clusters by teamspace...")
	user, driver := GetUserFromContext(c)

	teamspaceID := c.Param("id")
	if teamspaceID == "" {
		log.Println("Invalid teamspace ID")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
		return
	}
	id, err := primitive.ObjectIDFromHex(teamspaceID)
	if err != nil {
		log.Printf("Invalid teamspace ID %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
		return
	}

	teamspace := models.Teamspace{
		ID: id,
	}

	err = teamspace.Get(driver)
	if err != nil {
		log.Printf("Error getting teamspace %v", err)
		if utils.OnNotFoundError(err, "Teamspace") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Teamspace not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting teamspace"})
		}
		return
	}
	yes := teamspace.HasMember(driver, models.Member{UserID: user.ID.Hex()})
	if !yes {
		log.Printf("User %s is not a member of the teamspace", user.ID)
		c.JSON(http.StatusForbidden, gin.H{"message": "You are not a member of the teamspace"})
		return
	}

	cluster := models.Cluster{
		Teamspaces: []string{teamspaceID},
	}

	clusters, err := cluster.GetAllByTeamspace(driver)
	if err != nil {
		log.Printf("Error getting clusters %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting clusters"})
		}
		return
	}
	log.Println("Clusters retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"clusters": clusters, "size": len(clusters)})
}

func UserHasRightOnCluster(c *gin.Context, driver db.Driver, cluster models.Cluster, user models.User, roles []string) bool {

	// Check if the user is the creator of the cluster before updating (important if the cluster is not in a teamspace)
	err := cluster.GetByCreator(driver, user.ID)
	if err == nil {
		return true
	}
	if utils.OnNotFoundError(err, "Cluster") == nil {
		log.Printf("Error getting cluster %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting cluster"})
		return false
	}

	// or if the cluster is in a teamspace check if the user is a member of the teamspace and has the right permissions
	if len(cluster.Teamspaces) == 0 {
		log.Printf("Error getting cluster %v", err)
		if utils.OnNotFoundError(err, "Cluster") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cluster not found"})
			return false
		}
	} else {
		// if the cluster is not created by the current user, check if the user is a member of the teamspace
		for _, teamspaceID := range cluster.Teamspaces {
			t_id, err := primitive.ObjectIDFromHex(teamspaceID)
			if err != nil {
				log.Printf("Invalid teamspace ID %v", err)
				// c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
				// return false
				continue
			}
			teamspace := models.Teamspace{
				ID: t_id,
			}
			// Check if the user has the right to udpate or delete the teamspace
			yes, _, _ := MemberHasEnoughPrivilege(driver, roles, teamspace, user)
			if yes {
				return true
			}
		}
		log.Printf("User %s does not have the right to update or delete the cluster", user.ID)
		c.JSON(http.StatusForbidden, gin.H{"message": "You do not have the right to update or delete the cluster"})
		return false
	}
	return true
}

func setupCluster(driver db.Driver, clusterForm ClusterForm, user models.User) (models.Cluster, int, string) {
	cluster := models.Cluster{
		Name:        clusterForm.Name,
		Description: clusterForm.Description,
		Address:     clusterForm.Address,
		Port:        clusterForm.Port,
		CreatorID:   user.ID.Hex(),
		CreatedAt:   clusterForm.CreatedAt,
		Teamspaces:  clusterForm.Teamspaces,
	}
	var err error
	// if the cluster is global, it means that it is accessible by all the teamspaces the user is in
	if clusterForm.IsGlobal {
		cluster.Teamspaces = user.JoinedTeamspaceIDs
		t := models.Teamspace{
			CreatorID: user.ID.Hex(),
		}
		ts, err := t.GetAllByCreator(driver)
		if err != nil {
			log.Printf("Error getting teamspace %v", err)
			return models.Cluster{}, http.StatusInternalServerError, err.Error()
		}
		for _, ts := range ts {
			cluster.Teamspaces = append(cluster.Teamspaces, ts.ID.Hex())
		}
	}

	exp, err := GetTokenExpirationDate(clusterForm.Token)
	if err != nil {
		log.Printf("Error getting token expiration date: %v", err)
		return models.Cluster{}, http.StatusInternalServerError, err.Error()
	}
	cluster.ExpiryDate = exp
	token, err := generateClusterJWT(cluster, clusterForm.Token)
	if err != nil {
		log.Printf("Error generating cluster token: %v", err)
		return models.Cluster{}, http.StatusInternalServerError, err.Error()
	}
	cluster.Token = token
	return cluster, 0, ""
}

func generateClusterJWT(cluster models.Cluster, token string) (string, error) {
	claims := make(map[string]interface{})
	claims["sub"] = os.Getenv("KDI_JWT_SUB_FOR_K8S_API")
	claims["token"] = token
	claims["addr"] = cluster.Address
	claims["port"] = cluster.Port
	// TODO: same expiration date as the token
	claims["exp"] = cluster.ExpiryDate.Unix()

	return GenerateJWT(claims)
}

func clusterFormIsInValid(form ClusterForm) bool {
	return form.Name == "" || form.Address == "" || form.Token == "" ||
		(form.Type != models.TypeOpenshift && form.Type != models.TypeGKE && form.Type != models.TypeEKS && form.Type != models.TypeAKS)
}
