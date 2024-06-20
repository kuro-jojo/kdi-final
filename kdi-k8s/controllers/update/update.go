package update

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/kuro-jojo/kdi-k8s/models"
)

type UpdateForm struct {
	Name      string `json:"name" `
	Namespace string `json:"namespace" `
	Strategy  string `json:"strategy" binding:"required"`
	Container string `json:"container"`
	Image     string `json:"image" binding:"required"`
	Replicas  int32  `json:"replicas" binding:"required"`
	// Add more fields here depending on the strategy

	// For rolling update strategy
	MaxUnavailable string `json:"max_unavailable"` // The maximum number of pods that can be unavailable during the update process
	MaxSurge       string `json:"max_surge"`       // The maximum number of pods that can be scheduled above the desired number of pods
}

type DeploymentStatus struct {
	ReadyReplicas       int32    `json:"readyReplicas"`
	AvailableReplicas   int32    `json:"availableReplicas"`
	UnavailableReplicas int32    `json:"unavailableReplicas"`
	UpdatedReplicas     int32    `json:"updatedReplicas"`
	ReplicaFailures     []string `json:"replicaFailures"`
	Message             string   `json:"message"`
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
	updateForm.Name = c.Param("deployment")
	updateForm.Name = c.Param("deployment")

	switch updateForm.Strategy {
	case models.RollingUpdateStrategy:
		UpdateUsingRollingUpdateStrategy(c, updateForm)
	case models.RecreateStrategy:
		UpdateUsingRecreateStrategy(c, updateForm)
	// case models.ABTestingStrategy:
	// 	UpdateUsingABTestingStrategy(c, updateForm)
	// case models.CanaryStrategy:
	// 	UpdateUsingCanaryStrategy(c, updateForm)
	case models.BlueGreenStrategy:
		UpdateUsingBlueGreenStrategy(c, updateForm)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid strategy"})
	}
}

func isUpdateFormValid(form UpdateForm) bool {
	return form.Strategy != "" && form.Image != "" && form.Replicas > 0
}
