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

// TeamspaceForm is a struct that represents the form for creating a teamspace
type TeamspaceForm struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateTeamspace is a controller that creates a teamspace
func CreateTeamspace(c *gin.Context) {
	log.Println("Creating teamspace...")
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	var teamspaceForm TeamspaceForm
	if c.BindJSON(&teamspaceForm) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if teamspaceFormIsInValid(teamspaceForm) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	u, _ := c.Get("user")
	user := u.(models.User)

	var err error
	// check if user exists
	err = user.Get(driver)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "User not found"})
		return
	}

	teamspace := models.Teamspace{
		Name:        teamspaceForm.Name,
		Description: teamspaceForm.Description,
		CreatorID:   user.ID.Hex(),
		CreatedAt:   time.Now(),
	}

	err = teamspace.Create(driver)
	if err != nil {
		log.Printf("Error creating project %v", err)
		if er := utils.OnDuplicateKeyError(err, "Teamspace"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	// TODO: add creator to teamspace
	// c.Request.Form.Add("teamspace_id", teamspace.ID.Hex())
	// c.Request.Form.Add("user_id", user.ID.Hex())
	// c.Request.Form.Add("profile_name", "creator")
	// AddMemberToTeamspace(c)
	c.JSON(http.StatusCreated, gin.H{"message": "Teamspace created successfully"})
}

func GetTeamspacesByCreator(c *gin.Context) {
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	u, _ := c.Get("user")
	user := u.(models.User)

	teamspace := models.Teamspace{
		CreatorID: user.ID.Hex(),
	}
	teamspaces, err := teamspace.GetAllByCreator(driver)
	if err != nil {
		log.Printf("Error getting teamspaces %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting teamspaces"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teamspaces": teamspaces, "size": len(teamspaces)})
}

func GetTeamspace(c *gin.Context) {
	user, driver := GetUserFromContext(c)
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
		return
	}

	teamspace := models.Teamspace{
		ID:        objectID,
		CreatorID: user.ID.Hex(),
	}

	err = teamspace.Get(driver)

	if err != nil {
		log.Printf("Error getting teamspace %v", err)
		if er := utils.OnDuplicateKeyError(err, "Teamspace"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// check if user has the right to get the teamspace
	if user.ID.Hex() != teamspace.CreatorID {
		// Check if user is a member of the teamspace
		member := models.Member{
			UserID: user.ID.Hex(),
		}
		if !teamspace.HasMember(driver, member) {
			c.JSON(http.StatusForbidden, gin.H{"message": "User is not a member of the teamspace"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"teamspace": teamspace})
}

func teamspaceFormIsInValid(form TeamspaceForm) bool {
	return form.Name == ""
}
