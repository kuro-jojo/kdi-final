package controllers

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TokenEpirationDate = time.Hour * 24 * 1 // 1 days
)

type UserForm struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	log.Printf("Logging in user...")

	d, _ := c.Get("driver")
	driver := d.(db.Driver)
	var userForm UserForm

	if c.BindJSON(&userForm) != nil {
		log.Printf("Invalid form %v", c.BindJSON(&userForm))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if formIsInValid(userForm, true, false) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}

	var user models.User = models.User{
		Name:     userForm.Name,
		Email:    userForm.Email,
		Password: userForm.Password,
	}
	// Check if user exists
	err := user.GetByEmail(driver)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials"})
			return
		}
		log.Printf("Error while checking email %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while checking email"})
		return
	}
	// Check if password is correct
	if !models.CheckPasswordHash(userForm.Password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials"})
		return
	}

	token, err := generateUserAuthToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while generating token"})
		return
	}
	log.Printf("User logged successfully")
	c.JSON(http.StatusOK, gin.H{"message": "User logged successfully", "token": token})
}

func Register(c *gin.Context) {
	log.Printf("Registering new user...")

	d, _ := c.Get("driver")
	driver := d.(db.Driver)
	var userForm UserForm

	if c.BindJSON(&userForm) != nil {
		log.Printf("Invalid form %v", c.BindJSON(&userForm))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form"})
		return
	}

	if formIsInValid(userForm, false, false) {
		log.Printf("Invalid form %v", userForm)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid form fields"})
		return
	}
	var user models.User = models.User{
		Name:     userForm.Name,
		Email:    userForm.Email,
		Password: userForm.Password,
	}

	// Check if user esxists
	err := user.GetByEmail(driver)
	if err == nil {
		log.Printf("Email already used")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email already used"})
		return
	}

	if !strings.Contains(err.Error(), "not found") {
		log.Printf("Error while checking email %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while checking email"})
		return
	}

	// Create user
	err = user.Create(driver)
	if err != nil {
		log.Printf("Error while creating user %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while creating user"})
		return
	}

	log.Printf("User created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func RegisterWithMsal(c *gin.Context) {
	log.Printf("Registering new user with msal...")

	user, driver := GetUserFromContext(c)

	user.SignWith = "MSAL"

	// Check if user esxists
	err := user.GetByEmail(driver)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "User already exists"})
		return
	}

	if !strings.Contains(err.Error(), "not found") {
		log.Printf("Error while checking email %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while checking email"})
		return
	}

	// Create user
	err = user.Create(driver)
	if err != nil {
		log.Printf("Error while creating user %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error while creating user"})
		return
	}

	log.Printf("User created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func GetUser(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(models.User)

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetUserById(c *gin.Context) {
	_, driver := GetUserFromContext(c)
	id := c.Param("user_id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	user := models.User{
		ID: objectID,
	}

	err := user.Get(driver)

	if err != nil {
		log.Printf("Error getting user %v", err)
		if er := utils.OnDuplicateKeyError(err, "User"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}
	//c.JSON(http.StatusOK, gin.H{"message": message})
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func GetAllJoinedTeamspaces(c *gin.Context) {
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	u, _ := c.Get("user")
	user := u.(models.User)

	teamspaces, err := user.GetAllJoinedTeamspaces(driver)
	if err != nil {
		log.Printf("Error getting teamspaces %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error getting teamspaces"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"teamspaces": teamspaces, "size": len(teamspaces)})
}

func generateUserAuthToken(user models.User) (string, error) {
	claims := make(map[string]interface{})
	claims["sub"] = user.ID
	claims["exp"] = time.Now().Add(TokenEpirationDate).Unix()

	return GenerateJWT(claims)
}

func formIsInValid(u UserForm, ignoreName bool, ignorePassword bool) bool {
	b, err := regexp.MatchString(`^[a-zA-Z0-9_+&*-]+(?:\.[a-zA-Z0-9_+&*-]+)*@(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`, u.Email)
	return err != nil || !b || (u.Password == "" && !ignorePassword) || (u.Name == "" && !ignoreName)
}
