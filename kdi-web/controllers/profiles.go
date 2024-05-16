package controllers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
)

type ProfileForm struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

func CreateProfile(c *gin.Context) {
	log.Println("Creating new profile...")
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	var profileForm ProfileForm
	if c.BindJSON(&profileForm) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if profileFormIsInValid(profileForm) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	profile := models.Profile{
		Name:  profileForm.Name,
		Roles: profileForm.Roles,
	}
	rolesErr := profile.VerifyRoles()
	if len(rolesErr) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Undefined roles : " + strings.Join(rolesErr, ", ")})
		return
	}

	err := profile.Create(driver)
	if err != nil {
		log.Printf("Error adding new profile %v", err)
		if er := utils.OnDuplicateKeyError(err, "Profile"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

func GetProfiles(c *gin.Context) {
	log.Println("Getting profiles...")
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	profile := models.Profile{}
	profiles, err := profile.GetAll(driver)
	if err != nil {
		log.Printf("Error getting profiles %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	log.Println("Profiles retrieved successfully")
	c.JSON(http.StatusOK, gin.H{"profiles": profiles, "size": len(profiles)})
}

func GetDefinedRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"roles": models.GetRoles()})
}

func profileFormIsInValid(profileForm ProfileForm) bool {
	return profileForm.Name == "" || len(profileForm.Roles) == 0
}
