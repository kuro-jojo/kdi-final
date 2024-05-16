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

type NotificationForm struct {
	SenderID    string `json:"senderId"`
	TeamspaceID string `json:"teamspaceId"`
	Content     string `json:"content"`
	CreatedAt   string `json:"createdAt"`
}

func GetNotifications(c *gin.Context) {
	log.Println("Getting notifications...")

	user, driver := GetUserFromContext(c)
	notification := models.Notification{ID: user.ID}

	err := notification.Get(driver)
	if err != nil {
		if err = utils.OnNotFoundError(err, "Notification"); err == nil {
			log.Printf("Error getting notifications: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notification.Messages, "size": len(notification.Messages)})
}

func ReadNotification(c *gin.Context) {
	log.Println("Reading notification...")

	user, driver := GetUserFromContext(c)
	notificationForm := NotificationForm{}
	if err := c.ShouldBindJSON(&notificationForm); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}
	notification := models.Notification{ID: user.ID}
	createdAt, err := time.Parse(time.RFC3339, notificationForm.CreatedAt)
	if err != nil {
		log.Printf("Error parsing created_at: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid created_at format"})
		return
	}

	notitificationContent := models.NotificationContent{
		SenderID:    notificationForm.SenderID,
		TeamspaceID: notificationForm.TeamspaceID,
		CreatedAt:   createdAt,
		Content:     notificationForm.Content,
	}
	err = notification.Read(driver, notitificationContent)
	if err != nil {
		log.Printf("Error reading notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	log.Println("Notification read successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Notification read successfully"})
}

func DeleteNotifications(c *gin.Context) {
	log.Println("Deleting all notifications...")

	user, driver := GetUserFromContext(c)
	notification := models.Notification{ID: user.ID}

	err := notification.Delete(driver)
	if err != nil {
		log.Printf("Error deleting notifications: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}
	log.Println("Notifications deleted successfully")
	c.JSON(http.StatusOK, gin.H{"message": "Notifications deleted successfully"})
}

func SendNotificationToMember(c *gin.Context, userMember models.User, teamspace models.Teamspace, messageContent string) bool {
	user, driver := GetUserFromContext(c)

	notification := models.Notification{
		ID: userMember.ID,
	}

	err := notification.Get(driver)
	notificationContent := models.NotificationContent{
		SenderID:    user.ID.Hex(),
		TeamspaceID: teamspace.ID.Hex(),
		Content:     messageContent,
		CreatedAt:   time.Now(),
		WasRead:     false,
	}

	notification.Messages = append(notification.Messages, notificationContent)
	if err != nil {

		if err = utils.OnNotFoundError(err, "Notification"); err != nil {
			log.Printf("Notification section doesn't exist, creating it...")
			err = notification.Create(driver)
			if err != nil {
				log.Printf("Error creating notification: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
				return false
			}
		} else {
			log.Printf("Error getting notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return false
		}
	} else {

		log.Printf("Notification section exists, adding message to it...")
		err = notification.AddMessage(driver, notificationContent)
		if err != nil {
			log.Printf("Error adding message to notification: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return false
		}
	}

	log.Printf("Notification sent to %s", userMember.Email)
	return true
}

func SendNotificationToAllMembers(c *gin.Context, userMember models.User, teamspace models.Teamspace, driver db.Driver, messageContent string) bool {
	for _, m := range teamspace.Members {
		if m.UserID != userMember.ID.Hex() {
			m_id, err := primitive.ObjectIDFromHex(m.UserID)
			if err != nil {
				log.Printf("Error getting member ID: %v", err)
				continue
			}

			member := models.User{
				ID: m_id,
			}
			err = member.Get(driver)
			if err != nil {
				log.Printf("Error getting member: %v", err)
				continue
			}
			ok := SendNotificationToMember(c, member, teamspace, messageContent)
			if !ok {
				return false
			}
		}
	}
	return true
}
