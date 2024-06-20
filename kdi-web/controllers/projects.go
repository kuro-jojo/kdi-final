package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProjectForm struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TeamspaceID string `json:"teamspace_id"`
}

func CreateProject(c *gin.Context) {
	log.Println("Creating project...")

	code, message := createProject(c)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	log.Printf("Project created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Project created successfully"})
}

func GetProjectsByCreator(c *gin.Context) {
	log.Println("Listing all projects of the current user...")
	user, driver := GetUserFromContext(c)

	p := models.Project{
		CreatorID: user.ID.Hex(),
	}

	projects, err := p.GetAllByCreator(driver)
	if err != nil {
		log.Printf("Error getting projects %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects, "size": len(projects)})
}

func GetProjectsByTeamspace(c *gin.Context) {
	log.Println("Listing all projects in teamspace...")
	user, driver := GetUserFromContext(c)

	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
		return
	}

	teamspace := models.Teamspace{
		ID: objectID,
	}
	err = teamspace.Get(driver)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Teamspace not found"})
		return
	}

	// Vérifiez les privilèges de l'utilisateur

	// Vérifiez les privilèges de l'utilisateur
	ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.ListProjectsRole}, teamspace, user)
	if !ok {
		c.JSON(code, gin.H{"message": message})
		return
	}

	// Obtenir la liste des projets
	// Obtenir la liste des projets
	p := models.Project{
		TeamspaceID: teamspace.ID.Hex(),
	}

	projects, err := p.GetAllByTeamspace(driver)
	if err != nil {
		log.Printf("Error getting projects %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting projects"})
		return
	}

	// Construire la réponse JSON avec toutes les données nécessaires
	response := gin.H{
		"teamspace": teamspace,
		"projects":  projects,
		"size":      len(projects),
	}

	// Envoyer la réponse JSON
	c.JSON(http.StatusOK, response)
}

func GetProjectsOfJoinedTeamspaces(c *gin.Context) {
	// Récupérer le driver et l'utilisateur depuis le contexte
	user, driver := GetUserFromContext(c)

	// Obtenir la liste des teamspaces auxquels l'utilisateur a rejoint
	teamspaces, err := user.GetAllJoinedTeamspaces(driver)
	if err != nil {
		log.Printf("Error getting teamspaces %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting teamspaces"})
		return
	}

	allProjects := []models.Project{}

	// Pour chaque teamspace, obtenir les projets et les ajouter à la liste
	for _, teamspace := range teamspaces {
		p := models.Project{
			TeamspaceID: teamspace.ID.Hex(),
		}
		projects, err := p.GetAllByTeamspace(driver)
		if err != nil {
			log.Printf("Error getting projects %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting projects"})
			return
		}
		allProjects = append(allProjects, projects...)
	}

	response := gin.H{
		"projects": allProjects,
		"size":     len(allProjects),
	}

	// Envoyer la réponse JSON
	c.JSON(http.StatusOK, response)
}

func setupProject(c *gin.Context) (db.Driver, ProjectForm, models.User, int, string) {
	user, driver := GetUserFromContext(c)

	var projectForm ProjectForm
	if c.BindJSON(&projectForm) != nil {
		return nil, ProjectForm{}, models.User{}, http.StatusBadRequest, "Invalid form"
	}

	if projectFormIsInValid(projectForm) {
		return nil, ProjectForm{}, models.User{}, http.StatusBadRequest, "Invalid form fields"
	}

	return driver, projectForm, user, 0, ""
}

func createProject(c *gin.Context) (int, string) {
	driver, projectForm, user, code, message := setupProject(c)
	if code != 0 {
		return code, message
	}

	// Créez le projet en utilisant les données du formulaire
	// Créez le projet en utilisant les données du formulaire
	project := models.Project{
		Name:        projectForm.Name,
		Description: projectForm.Description,
		CreatorID:   user.ID.Hex(),
		CreatedAt:   time.Now(),
	}

	// Vérifiez si TeamspaceID est fourni dans le formulaire
	if projectForm.TeamspaceID != "" {
		// Si TeamspaceID est fourni, vérifiez s'il est valide
		t_id, err := primitive.ObjectIDFromHex(projectForm.TeamspaceID)
		if err != nil {
			return http.StatusBadRequest, "Invalid teamspace ID"
		}

		// Récupérez les détails du Teamspace
		teamspace := models.Teamspace{ID: t_id}
		err = teamspace.Get(driver)
		if err != nil {
			return http.StatusBadRequest, "Teamspace not found"
		}

		// Vérifiez si l'utilisateur a les autorisations nécessaires pour accéder à ce Teamspace
		ok, code, message := MemberHasEnoughPrivilege(driver, []string{models.CreateProjectRole}, teamspace, user)
		if !ok {
			return code, message
		}

		// Associez le projet au Teamspace
		project.TeamspaceID = t_id.Hex()

	}

	// Créez le projet dans la base de données
	err := project.Create(driver)
	if err != nil {
		log.Printf("Error creating project: %v", err)
		if er := utils.OnDuplicateKeyError(err, "Project"); er != nil {
			return http.StatusConflict, er.Error()
		} else {
			return http.StatusInternalServerError, err.Error()
		}
	}

	return 0, ""
}

func GetProject(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	project := models.Project{
		ID:        objectID,
		CreatorID: user.ID.Hex(),
	}

	err := project.Get(driver)

	if err != nil {
		log.Printf("Error getting project %v", err)
		if utils.OnNotFoundError(err, "Project") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": message})
	c.JSON(http.StatusOK, gin.H{"project": project})
}

func DeleteProject(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("id")

	projectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project := models.Project{
		ID:        projectID,
		CreatorID: user.ID.Hex(),
	}

	err = project.Delete(driver)
	if err != nil {
		if utils.OnNotFoundError(err, "Project") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	// Répondre avec un message indiquant que le projet a été supprimé avec succès
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func projectFormIsInValid(form ProjectForm) bool {
	return form.Name == ""
}

func UpdateProject(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("id")
	// Vérification de la validité de l'ID du projet
	projectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	updatedProject := models.Project{
		ID:        projectID,
		CreatorID: user.ID.Hex(),
	}
	if err := c.BindJSON(&updatedProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project data"})
		return
	}
	if err := updatedProject.Update(driver); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Updated project": updatedProject})
}
