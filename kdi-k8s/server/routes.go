package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-k8s/auth"
	controllersdeployments "github.com/kuro-jojo/kdi-k8s/controllers/deployments"
	controllersnamespaces "github.com/kuro-jojo/kdi-k8s/controllers/namespaces"
	controllersupdate "github.com/kuro-jojo/kdi-k8s/controllers/update"
	controllersfiles "github.com/kuro-jojo/kdi-k8s/files/controllers"
)

func authMiddleware(c *gin.Context) {
	auth.AuthenticateToCluster(c)
}

// SetupRoutes sets up the routes for the kubernetes API
func SetupRoutes(group *gin.RouterGroup) {

	authenticated := group.Group("")

	authenticated.Use(authMiddleware)
	{
		authenticated.GET("/auth", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		resources := authenticated.Group("/resources/deployments")
		{
			resources.POST("/with-yaml", controllersfiles.CreateDeployment)
			// resources.POST("/with-helm",files.CreateDeployment)
		}
		authenticated.POST("/resources/services/with-yaml", controllersfiles.CreateService)
		authenticated.POST("/resources/with-yaml", controllersfiles.CreateMultipleRessources)

		// resources bounded to a namespace
		namespaces := authenticated.Group("/resources/namespaces")
		{
			namespaces.GET("", controllersnamespaces.GetNamespaces)
			// namespaces.GET("/:namespace", controllersdeployments.GetNamespace)
			// namespaces.GET(":namespace/deployments", controllersdeployments.GetDeploymentsInNamespace)
			namespaces.GET(":namespace/deployments/:deployment", controllersdeployments.GetDeploymentInNamespace)

			namespaces.PATCH(":namespace/deployments/:deployment", controllersupdate.UpdateDeployment)
		}
	}
}
