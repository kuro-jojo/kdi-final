package controllers

import (
	"log"
	"net/http"

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
