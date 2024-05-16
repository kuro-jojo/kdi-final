package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
)

func DbMiddleware(driver db.Driver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if driver == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Database connection error"})
			return
		}
		c.Set("driver", driver)
		c.Next()
	}
}
