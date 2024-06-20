package update

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/utils"
	"k8s.io/client-go/util/retry"
)

// This file contains the base kubernetes strategies (rolling update and recreate) for updating a deployment

func UpdateUsingRollingUpdateStrategy(c *gin.Context, updateForm UpdateForm) {
	log.Println("Updating deployment using rolling update strategy...")

	// check if maxUnavailable is a number or a percentage
	if updateForm.MaxUnavailable != "" && !isNumber(updateForm.MaxUnavailable) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "MaxUnavailable must be a number or a percentage"})
		return
	}
	// check if maxSurge is a number or a percentage
	if updateForm.MaxSurge != "" && !isNumber(updateForm.MaxSurge) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "MaxSurge must be a number or a percentage"})
		return
	}

	if isZero(updateForm.MaxUnavailable) && isZero(updateForm.MaxSurge) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "MaxSurge cannot be 0 if MaxUnavailable is 0 or vice versa"})
		return
	}

	// set default values for maxUnavailable and maxSurge if not provided

	if updateForm.MaxUnavailable == "" {
		updateForm.MaxUnavailable = "25%"
	} else if isZero(updateForm.MaxUnavailable) {
		updateForm.MaxUnavailable = "0%"
	}

	if updateForm.MaxSurge == "" {
		updateForm.MaxSurge = "25%"
	} else if isZero(updateForm.MaxSurge) {
		updateForm.MaxSurge = "0%"
	}

	ok := updateUsingK8sStrategy(c, updateForm, v1.RollingUpdateDeploymentStrategyType)
	if !ok {
		return
	}

	log.Printf("Deployment %s updated successfully", updateForm.Name)
	c.JSON(http.StatusOK, gin.H{"message": "Deployment updated successfully"})
}

func UpdateUsingRecreateStrategy(c *gin.Context, updateForm UpdateForm) {
	log.Println("Updating deployment using recreate strategy...")

	ok := updateUsingK8sStrategy(c, updateForm, v1.RecreateDeploymentStrategyType)
	if !ok {
		return
	}

	log.Printf("Deployment %s updated successfully", updateForm.Name)
	c.JSON(http.StatusOK, gin.H{"message": "Deployment updated successfully"})
}

func updateUsingK8sStrategy(c *gin.Context, updateForm UpdateForm, strategy v1.DeploymentStrategyType) bool {

	clientset := utils.GetClientSet(c)
	deployment := clientset.AppsV1().Deployments(updateForm.Namespace)

	// Retrieve the latest version of Deployment before attempting update
	// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		result, err := deployment.Get(context.TODO(), updateForm.Name, metav1.GetOptions{})
		if err != nil && utils.IsNotFoundError(err.Error()) {
			return fmt.Errorf("deployment %s not found in namespace %s", updateForm.Name, updateForm.Namespace)
		}
		result.Spec.Replicas = &updateForm.Replicas
		result.Spec.Strategy.Type = strategy
		if strategy == v1.RecreateDeploymentStrategyType {
			result.Spec.Strategy.RollingUpdate = nil
		} else if strategy == v1.RollingUpdateDeploymentStrategyType {
			result.Spec.Strategy.RollingUpdate = &v1.RollingUpdateDeployment{
				MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: updateForm.MaxUnavailable},
				MaxSurge:       &intstr.IntOrString{Type: intstr.String, StrVal: updateForm.MaxSurge},
			}
		}
		result.Spec.Template.Spec.Containers[0].Image = updateForm.Image
		_, updateErr := deployment.Update(context.TODO(), result, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("failed to update deployment %s: %v", updateForm.Name, updateErr)
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return false
	}
	return true
}

func isNumber(val string) bool {
	if val != "" {
		if strings.HasSuffix(val, "%") {
			_, err := strconv.Atoi(strings.Replace(val, "%", "", -1))

			return err == nil
		} else {
			_, err := strconv.Atoi(val)
			return err == nil
		}
	}
	return false
}

func isZero(val string) bool {
	if val == "0" || val == "0%" {
		return true
	}
	return false
}
