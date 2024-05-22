package server

import (
	"github.com/kuro-jojo/kdi-web/controllers"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupRoutes sets up the routes for the web API
func SetupRoutes(group *gin.RouterGroup, driver db.Driver, msalAuth middlewares.MsalWebAuth) {

	// all routes except login and register require authentication
	route := group.Group("")
	route.Use(middlewares.DbMiddleware(driver))

	route.POST("login", controllers.Login)
	route.POST("register", controllers.Register)

	// this route is for checking the health of the server (if it is up)
	route.GET("health", controllers.Health)

	// all routes below require authentication
	authenticatedRoute := route.Group("")
	authenticatedRoute.Use(middlewares.AuthMiddleware(msalAuth))
	// this route is for registering with msal after the user has been authenticated with msal
	authenticatedRoute.POST("register/msal", controllers.RegisterWithMsal)

	dashboard := authenticatedRoute.Group("dashboard")
	{
		users := dashboard.Group("users")
		{
			users.GET("", controllers.GetUser)
			users.GET("notifications", controllers.GetNotifications)
			users.PATCH("notifications", controllers.ReadNotification)
			users.DELETE("notifications", controllers.DeleteNotifications)
			users.GET(":user_id", controllers.GetUserById)
		}

		projects := dashboard.Group("projects")
		{
			projects.GET("owned", controllers.GetProjectsByCreator)
			projects.GET("joinedTeamspaces", controllers.GetProjectsOfJoinedTeamspaces)
			projects.POST("", controllers.CreateProject)
			projects.GET(":id", controllers.GetProject)
			projects.PATCH(":id", controllers.UpdateProject)
			projects.DELETE(":id", controllers.DeleteProject)
		}

		teamspaces := dashboard.Group("teamspaces")
		{
			teamspaces.POST("", controllers.CreateTeamspace)
			teamspaces.GET("owned", controllers.GetTeamspacesByCreator)
			teamspaces.GET("joined", controllers.GetAllJoinedTeamspaces)
			teamspaces.GET(":id", controllers.GetTeamspace)

			teamspaces.GET(":id/projects", controllers.GetProjectsByTeamspace)

			teamspaces.PATCH(":id/members", controllers.AddMemberToTeamspace)
			teamspaces.DELETE(":id/members/:memberId", controllers.RemoveMemberFromTeamspace)
			teamspaces.PATCH(":id/members/:memberId", controllers.UpdateMemberInTeamspace)
		}

		profiles := dashboard.Group("profiles")
		{
			profiles.POST("", controllers.CreateProfile)
			profiles.GET("", controllers.GetProfiles)
			profiles.GET("roles", controllers.GetDefinedRoles)
		}

		clusters := dashboard.Group("clusters")
		{
			clusters.POST("", controllers.AddCluster)
			clusters.GET("owned", controllers.GetClustersByCreator)
			clusters.GET("teamspaces", controllers.GetClustersByTeamspace)
			clusters.GET(":id", controllers.GetClusterByIDAndCreator)
			clusters.PATCH(":id", controllers.UpdateCluster)
			clusters.DELETE(":id", controllers.DeleteCluster)

			clusters.GET(":id/environments", controllers.GetEnvironmentsByCluster)
		}

		environments := dashboard.Group("environments")
		{
			environments.POST("", controllers.CreateEnvironment)
			environments.GET("", controllers.GetEnvironments)
			environments.GET(":e_id", controllers.GetEnvironment)
			environments.GET("projects/:project_id", controllers.GetEnvironmentsByProject)

			microservices := environments.Group(":e_id/microservices")
			{
				// microservices.POST("", controllers.CreateMicroservice)
				microservices.GET("", controllers.GetMicroservicesByEnvironment)
				microservices.POST("with-yaml", controllers.CreateMicroserviceWithYaml)
				microservices.GET(":m_id", controllers.GetMicroserviceByEnvironment)
			}
		}
	}
}
