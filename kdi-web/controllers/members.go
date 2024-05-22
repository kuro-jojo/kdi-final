package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MemberForm struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ProfileID string `json:"profile"`
}

const (
	AddingMemberNotificationContent = "%s added you to the teamspace %s as '%s'"
	NewMemberAddedInformation       = "%s added %s to the teamspace %s as '%s'"

	RetiringMemberNotificationContent = "%s removed you from the teamspace %s"
	MemberRemovedInformation          = "%s removed %s from the teamspace %s"

	MemberProfileUpdateNotificationContent = "%s updated your profile to '%s' in the teamspace %s"
	MemberProfileUpdatedInformation        = "%s updated %s's profile to '%s' in the teamspace %s"
)

func AddMemberToTeamspace(c *gin.Context) {
	log.Println("Adding member to teamspace...")

	driver, teamspace, member, userMember, code, message := setupMember(c, []string{models.AddMemberRole}, false, false)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	member.JoinDate = time.Now()

	err := teamspace.AddMember(driver, member)
	if err != nil {
		log.Printf("Error adding member to the teamspace: %v", err)
		if er := utils.OnDuplicateKeyError(err, "Member"); er != nil {
			c.JSON(http.StatusConflict, gin.H{"message": "User's already member of the teamspace"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// add the teamspace to the user's joined teamspaces

	err = userMember.AddToTeamspace(driver, teamspace)
	if err != nil {
		log.Printf("Error adding teamspace to user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})

		// Rollback
		err = teamspace.RemoveMember(driver, member)
		if err != nil {
			log.Printf("Error rolling back: %v", err)
		}
		return
	}
	// Send notification to the user
	user, _ := GetUserFromContext(c)
	messageContent := fmt.Sprintf(AddingMemberNotificationContent, user.Name, teamspace.Name, member.ProfileName)
	ok := SendNotificationToMember(c, userMember, teamspace, messageContent)
	if !ok {
		return
	}

	// Send notification to all members of the teamspace
	log.Println("Sending notification to all members of the teamspace...")
	messageContent = fmt.Sprintf(NewMemberAddedInformation, user.Name, userMember.Name, teamspace.Name, member.ProfileName)
	ok = SendNotificationToAllMembers(c, userMember, teamspace, driver, messageContent)
	if !ok {
		return
	}
	log.Println("Member added to teamspace successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Member added to teamspace successfully"})
}

func UpdateMemberInTeamspace(c *gin.Context) {
	log.Println("Updating member in teamspace...")
	driver, teamspace, member, userMember, code, message := setupMember(c, []string{models.UpdateMemberRole}, false, true)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	err := teamspace.UpdateMember(driver, member)
	if err != nil {
		log.Printf("Error updating member to the teamspace: %v", err)
		if er := utils.OnSameValueError(err, "profile"); er != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// Send notification to the user
	user, _ := GetUserFromContext(c)
	messageContent := fmt.Sprintf(MemberProfileUpdateNotificationContent, user.Name, member.ProfileName, teamspace.Name)
	ok := SendNotificationToMember(c, userMember, teamspace, messageContent)
	if !ok {
		return
	}

	// Send notification to all members of the teamspace
	log.Println("Sending notification to all members of the teamspace...")
	messageContent = fmt.Sprintf(MemberProfileUpdatedInformation, user.Name, userMember.Name, member.ProfileName, teamspace.Name)
	ok = SendNotificationToAllMembers(c, userMember, teamspace, driver, messageContent)
	if !ok {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Member updated in teamspace successfully"})
}

func RemoveMemberFromTeamspace(c *gin.Context) {
	log.Println("Removing member from teamspace...")
	driver, teamspace, member, userMember, code, message := setupMember(c, []string{models.RemoveMemberRole}, true, false)
	if code != 0 {
		c.JSON(code, gin.H{"message": message})
		return
	}

	err := teamspace.RemoveMember(driver, member)
	if err != nil {
		log.Printf("Error removing member to the teamspace: %v", err)
		if er := utils.OnNotFoundError(err, "Member"); er != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": er.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	// remove the teamspace from the user's joined teamspaces

	err = userMember.RemoveFromTeamspace(driver, teamspace)
	if err != nil {
		log.Printf("Error adding teamspace to user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})

		// Rollback
		err = teamspace.RemoveMember(driver, member)
		if err != nil {
			log.Printf("Error rolling back: %v", err)
		}
		return
	}

	// Send notification to the user
	user, _ := GetUserFromContext(c)
	messageContent := fmt.Sprintf(RetiringMemberNotificationContent, user.Name, teamspace.Name)
	ok := SendNotificationToMember(c, userMember, teamspace, messageContent)
	if !ok {
		return
	}

	// Send notification to all members of the teamspace
	log.Println("Sending notification to all members of the teamspace...")
	messageContent = fmt.Sprintf(MemberRemovedInformation, user.Name, userMember.Name, teamspace.Name)
	ok = SendNotificationToAllMembers(c, userMember, teamspace, driver, messageContent)
	if !ok {
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Member remove from teamspace successfully"})
}

// GetMembersByTeamspace returns all members of a teamspace
func GetMembersByTeamspace(c *gin.Context) {
	log.Println("Getting members of a teamspace...")
	user, driver := GetUserFromContext(c)

	teamspaceID := c.Param("teamspace_id")
	if teamspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid teamspace ID"})
		return
	}
	id, err := primitive.ObjectIDFromHex(teamspaceID)
	if err != nil {
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
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	yes, code, message := MemberHasEnoughPrivilege(driver, []string{models.ListMembersRole}, teamspace, user)
	if !yes {
		c.JSON(code, gin.H{"message": message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": teamspace.Members, "size": len(teamspace.Members)})
}

// setupMember is a helper function that sets up the member addition or removal (or else) to a teamspace
func setupMember(c *gin.Context, roles []string, isDeletion bool, isUpdate bool) (db.Driver, models.Teamspace, models.Member, models.User, int, string) {
	user, driver := GetUserFromContext(c)

	var memberForm MemberForm
	if !isDeletion {
		if c.BindJSON(&memberForm) != nil {
			log.Println("Error binding JSON : Invalid form")
			return nil, models.Teamspace{}, models.Member{}, models.User{}, http.StatusBadRequest, "Invalid form"
		}
		if (isUpdate && memberForm.ProfileID == "") || (!isUpdate && memberFormIsInValid(memberForm)) {
			log.Println("Error with the form: Invalid form values")
			return nil, models.Teamspace{}, models.Member{}, models.User{}, http.StatusBadRequest, "Invalid form"
		}
	}
	if isDeletion || isUpdate {
		memberForm.UserID = c.Param("memberId")
	}

	t_id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Println("Error with the form: Invalid teamspace ID")
		return nil, models.Teamspace{}, models.Member{}, models.User{}, http.StatusBadRequest, "Invalid teamspace ID"
	}

	teamspace := models.Teamspace{
		ID: t_id,
	}

	err = teamspace.Get(driver)
	if err != nil {
		log.Printf("Error getting teamspace %v", err)
		if utils.OnNotFoundError(err, "Teamspace") != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Teamspace not found"})
			return nil, models.Teamspace{}, models.Member{}, models.User{}, http.StatusBadRequest, "Teamspace not found"
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return nil, models.Teamspace{}, models.Member{}, models.User{}, http.StatusInternalServerError, err.Error()
		}
	}
	// Check if user has the right to add member to teamspace

	yes, code, message := MemberHasEnoughPrivilege(driver, roles, teamspace, user)
	if !yes {
		return driver, teamspace, models.Member{}, models.User{}, code, message
	}

	userMember := models.User{
		Email: memberForm.Email,
	}

	if memberForm.UserID == "" {
		err := userMember.GetByEmail(driver)
		if err != nil {
			log.Printf("Error getting user by email: %v", err)
			return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusBadRequest, "User not found"
		}
		memberForm.UserID = userMember.ID.Hex()
	} else {
		u_id, err := primitive.ObjectIDFromHex(memberForm.UserID)
		if err != nil {
			return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusBadRequest, "Invalid user ID"
		}
		userMember.ID = u_id
		err = userMember.Get(driver)
		if err != nil {
			log.Printf("Error getting user: %v", err)
			return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusBadRequest, "User not found"
		}
	}

	if userMember.ID.Hex() == user.ID.Hex() {
		return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusConflict, "User can't add himself to the teamspace"
	}
	member := models.Member{
		UserID: userMember.ID.Hex(),
		Name:   userMember.Name,
		Email:  userMember.Email,
	}
	if !isDeletion {
		p_id, err := primitive.ObjectIDFromHex(memberForm.ProfileID)
		if err != nil {
			log.Printf("Error getting profile ID: %v", err)
			return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusBadRequest, "Invalid profile ID"
		}

		profile := models.Profile{
			ID: p_id,
		}
		err = profile.Get(driver)
		if err != nil {
			log.Printf("Error getting profile: %v", err)
			return nil, models.Teamspace{}, models.Member{}, userMember, http.StatusBadRequest, "Profile not found"
		}
		member.ProfileName = profile.Name
	}

	return driver, teamspace, member, userMember, 0, ""
}

// MemberHasEnoughPrivilege checks if the user has enough privilege to do an action in a teamspace
func MemberHasEnoughPrivilege(driver db.Driver, roles []string, teamspace models.Teamspace, user models.User) (bool, int, string) {
	// Bypass if user is the creator
	if teamspace.CreatorID != user.ID.Hex() {
		// 1. Get all profiles that have the desired roles
		// 2. User must be a member of the team
		// 3. Must have a profile that has an role
		p := models.Profile{}
		profilesWithRole, err := p.GetAllByRoles(driver, roles)

		if err != nil {
			log.Printf("Error getting profiles with %v role: %v", roles, err)
			return false, http.StatusInternalServerError, "Internal server error"
		}

		profileNames := utils.ModelArrayToStringArray[models.Profile](profilesWithRole, func(p models.Profile) string {
			return p.Name
		})

		if !teamspace.HasMemberWithProfile(driver, user.ID.Hex(), profileNames) {
			log.Printf("User hasn't the right to %s in the teamspace", strings.Join(roles, ", "))
			return false, http.StatusForbidden, fmt.Sprintf("User hasn't the right to %s in the teamspace", strings.Join(roles, ", "))
		}
	}
	return true, 0, ""
}

func memberFormIsInValid(memberForm MemberForm) bool {
	return (memberForm.UserID == "" && memberForm.Email == "") || memberForm.ProfileID == ""
}
