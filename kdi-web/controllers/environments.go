package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnvironmentForm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ClusterID   string `json:"clusterId"`
	ProjectID   string `json:"projectId"`
}

func CreateEnvironment(c *gin.Context) {
	log.Println("Creating environment...")

	user, driver := GetUserFromContext(c)

	var environmentForm EnvironmentForm
	if c.BindJSON(&environmentForm) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if EnvironmentFormIsInValid(environmentForm) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields. Name, Cluster and Project are required fields"})
		return
	}

	var err error
	// check if user exists
	err = user.Get(driver)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User not found"})
		return
	}

	environment := models.Environment{
		Name:        environmentForm.Name,
		Description: environmentForm.Description,
		//ClusterID:   environmentForm.ClusterID,
	}

	cluster_id, err := primitive.ObjectIDFromHex(environmentForm.ClusterID)
	if err != nil {
		log.Println("Error creating environment: Invalid Cluster ID")
		c.JSON(http.StatusBadRequest, gin.H{"Invalid Cluster ID": err.Error()})
		return
	}

	cluster := models.Cluster{
		ID: cluster_id,
	}
	err = cluster.Get(driver)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Cluster not found": err.Error()})
	}
	// user must be a member of the teamspace and have access to create projects
	//ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.CREATE_PROJECT}, teamspace, user)
	//if !ok {
	//return code, message
	//}
	environment.ClusterID = cluster_id.Hex()
	if environmentForm.ProjectID != "" {
		project_id, err := primitive.ObjectIDFromHex(environmentForm.ProjectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Invalid Project ID": err.Error()})
		}

		project := models.Project{
			ID: project_id,
		}
		err = project.Get(driver)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Project not found": err.Error()})
		}
		// user must be a member of the teamspace and have access to create projects
		//ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.CREATE_PROJECT}, teamspace, user)
		//if !ok {
		//return code, message
		//}
		environment.ProjectID = project_id.Hex()
	}

	err = environment.Create(driver)
	if err != nil {
		log.Printf("Error creating environment %v", err)
		if er := utils.OnDuplicateKeyError(err, "Environment"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Environment created successfully"})

}

func EnvironmentFormIsInValid(form EnvironmentForm) bool {
	return form.Name == "" || form.ProjectID == "" || form.ClusterID == ""
}

func GetEnvironmentsByCluster(c *gin.Context) {
	log.Println("Listing all environments in a cluster...")
	_, driver := GetUserFromContext(c)

	var envForm EnvironmentForm
	if c.BindJSON(&envForm) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	env_id, err := primitive.ObjectIDFromHex(envForm.ClusterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid cluster ID"})
		return
	}

	cluster := models.Cluster{
		ID: env_id,
	}
	err = cluster.Get(driver)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cluster not found"})
		return
	}
	// user must be a member of the teamspace and have access to view projects
	//ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.LIST_PROJECTS}, cluster, user)
	//if !ok {
	//c.JSON(code, gin.H{"message": message})
	//return
	//}

	env := models.Environment{
		ClusterID: cluster.ID.Hex(),
	}

	environments, err := env.GetAllByCluster(driver)
	if err != nil {
		log.Printf("Error getting projects %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environments": environments, "size": len(environments)})
}

func GetEnvironments(c *gin.Context) {
	log.Println("Listing all environments...")
	_, driver := GetUserFromContext(c)
	//defer driver.Close()

	environment := models.Environment{}

	environments, err := environment.GetAll(driver)
	if err != nil {
		log.Printf("Error getting environments %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environments": environments, "size": len(environments)})
}

func GetEnvironment(c *gin.Context) {
	_, driver := GetUserFromContext(c)
	id := c.Param("e_id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	env := models.Environment{
		ID: objectID,
		//CreatorID: user.ID.Hex(),
	}

	err := env.Get(driver)

	if err != nil {
		log.Printf("Error getting environment %v", err)
		if er := utils.OnDuplicateKeyError(err, "Environment"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": message})
	c.JSON(http.StatusOK, gin.H{"environment": env})
}

/*func DeleteEnvironment(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("e_id")

	// Obtenir l'environnement
	envID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid environment ID"})
		return
	}

	env := models.Environment{
		ID: envID,
	}

	err = env.Get(driver)
	if err != nil {
		log.Printf("Error getting environment %v", err)
		if er := utils.OnDuplicateKeyError(err, "Environment"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Récupérer l'ID du projet à partir de l'environnement
	projectID := env.ProjectID
	projectObjectID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project := models.Project{
		ID: projectObjectID,
	}

	err = project.Get(driver)

	if err != nil {
		log.Printf("Error getting project %v", err)
		if er := utils.OnDuplicateKeyError(err, "Project"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Récupérer l'ID de la teamspace à partir du projet
	teamID := project.TeamspaceID
	teamObjectID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	team := models.Teamspace{
		ID: teamObjectID,
	}

	err = team.Get(driver)

	if err != nil {
		log.Printf("Error getting team %v", err)
		if er := utils.OnDuplicateKeyError(err, "Team"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Vérifiez si l'utilisateur a les autorisations nécessaires pour accéder à ce Teamspace
	ok, _, _ := MemberHasEnoughPrivilege(driver, []string{models.DeleteEnvironmentRole}, team, user)
	if !ok {
		return
	}

	if err := env.Delete(driver); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete environment"})
		return
	}

	// Répondre avec un message indiquant que le projet a été supprimé avec succès
	c.JSON(http.StatusOK, gin.H{"message": "Environment deleted successfully"})

}

func DeleteEnvironment(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("e_id")

	envID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid environment ID"})
		return
	}

	environment := models.Environment{
		ID: envID,
	}
	project_id := environment.ProjectID
	projectID, err := primitive.ObjectIDFromHex(project_id)
	if err != nil {
		log.Println("Project ID:", projectID.Hex())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	} else {
		// Log projectID
		log.Println("Project ID:", projectID.Hex())
		c.JSON(http.StatusOK, gin.H{"success": projectID})
	}

	projet := models.Project{
		ID: projectID,
	}
	teamspace_id := projet.TeamspaceID
	teamspaceID, err := primitive.ObjectIDFromHex(teamspace_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	} else {
		// Log teamspaceID
		log.Println("Teamspace ID:", teamspaceID.Hex())
		c.JSON(http.StatusOK, gin.H{"success": teamspaceID})
	}

	teamspace := models.Teamspace{
		ID: teamspaceID,
	}

	// Vérifiez si l'utilisateur a les autorisations nécessaires pour accéder à ce Teamspace
	ok, _, _ := MemberHasEnoughPrivilege(driver, []string{models.DeleteEnvironmentRole}, teamspace, user)
	if !ok {
		return
	}

	if err := environment.Delete(driver); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete environment"})
		return
	}

	// Répondre avec un message indiquant que le projet a été supprimé avec succès
	c.JSON(http.StatusOK, gin.H{"message": "Environment deleted successfully"})
}*/

func UpdateEnvironment(c *gin.Context) {
	_, driver := GetUserFromContext(c)
	id := c.Param("e_id")
	// Vérification de la validité de l'ID du projet
	envID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid environment ID"})
		return
	}
	updatedEnvironment := models.Environment{
		ID: envID,
	}
	if err := c.BindJSON(&updatedEnvironment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid environment data"})
		return
	}
	if err := updatedEnvironment.Update(driver); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update environment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Updated project": updatedEnvironment})
}

/* GetClustersByCreator returns all environments of the current user
func GetEnvironmentsByCreator(c *gin.Context) {
	log.Println("Listing all environments of the current user...")
	user, driver := GetUserFromContext(c)

	environment := models.Environment{
		CreatorID: user.ID.Hex(),
	}

	environments, err := environment.GetAllByCreator(driver)
	if err != nil {
		log.Printf("Error getting environment %v", err)
		if utils.OnNotFoundError(err, "Environment") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Environment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environment"})
		}
		return
	}
	log.Println("Environments retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"environments": environments, "size": len(environments)})
}*/

func GetEnvironmentsByProject(c *gin.Context) {
	log.Println("Listing all environments associated to a project...")
	_, driver := GetUserFromContext(c)

	id := c.Param("project_id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid project ID"})
		return
	}

	project := models.Project{
		ID: objectID,
	}
	err = project.Get(driver)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Project not found"})
		return
	}

	/* Vérifiez les privilèges de l'utilisateur
	ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.ListProjectsRole}, project, user)
	if !ok {
		c.JSON(code, gin.H{"message": message})
		return
	}*/

	// Obtenir la liste des projets
	e := models.Environment{
		ProjectID: project.ID.Hex(),
	}

	environments, err := e.GetAllByProject(driver)
	if err != nil {
		log.Printf("Error getting environments %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting environments"})
		return
	}

	// Construire la réponse JSON avec toutes les données nécessaires
	response := gin.H{
		"project":      project,
		"environments": environments,
		"size":         len(environments),
	}

	// Envoyer la réponse JSON
	c.JSON(http.StatusOK, response)
}

/*func getTokenInfo(token string) {
	claims := jwt.MapClaims{}
	token, _ := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("<YOUR VERIFICATION KEY>"), nil
	})

	// do something with decoded claims
	for key, val := range claims {
		fmt.Printf("Key: %v, value: %v\n", key, val)
	}
}*/
