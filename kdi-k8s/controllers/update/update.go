package update

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const (
	RollingUpdateStrategy = "rolling-update"
	RecreateStrategy      = "recreate"
	ABTestingStrategy     = "ab-testing"
	CanaryStrategy        = "canary"
	BlueGreenStrategy     = "blue-green"
)

type UpdateForm struct {
	DeploymentName string `json:"deployment_name" `
	Namespace      string `json:"namespace" `
	Strategy       string `json:"strategy" binding:"required"`
	Container      string `json:"container"`
	Image          string `json:"image" binding:"required"`
	Replicas       int32  `json:"replicas" binding:"required"`
	// Add more fields here depending on the strategy

	// For rolling update strategy
	MaxUnavailable        string  `json:"max_unavailable"` // The maximum number of pods that can be unavailable during the update process
	MaxSurge              string  `json:"max_surge"` // The maximum number of pods that can be scheduled above the desired number of pods
}

func UpdateDeployment(c *gin.Context) {
	var updateForm UpdateForm

	if c.ShouldBindBodyWith(&updateForm, binding.JSON) != nil {
		log.Printf("Invalid form %v", c.ShouldBindBodyWith(&updateForm, binding.JSON).Error())
		message := "Invalid form"
		if !isUpdateFormValid(updateForm) {
			message += " - Please provide at least deployment's image, replicas and strategy used"
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": message})
		return
	}
	updateForm.Namespace = c.Param("namespace")
	updateForm.DeploymentName = c.Param("deployment")

	switch updateForm.Strategy {
	case RollingUpdateStrategy:
		UpdateUsingRollingUpdateStrategy(c, updateForm)
	case RecreateStrategy:
		UpdateUsingRecreateStrategy(c, updateForm)
	// case ABTestingStrategy:
	// 	UpdateUsingABTestingStrategy(c, updateForm)
	// case CanaryStrategy:
	// 	UpdateUsingCanaryStrategy(c, updateForm)
	// case BlueGreenStrategy:
	// 	UpdateUsingBlueGreenStrategy(c, updateForm)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid strategy"})
	}
}

func isUpdateFormValid(form UpdateForm) bool {
	return form.Strategy != "" && form.Image != "" && form.Replicas > 0
}
