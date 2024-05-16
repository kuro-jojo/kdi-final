package utils

import (
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/kubernetes"
)

func GetMostSeenCode(httpResps map[int][]string) int {
	most := -1
	mv := 0
	for k, v := range httpResps {
		if len(v) > mv {
			most = k
			mv = len(v)
		}
	}
	return most
}

// GetClientSet returns the clientset from the context of the request
func GetClientSet(c *gin.Context) *kubernetes.Clientset {
	if cs, ok := c.Get("clientset"); ok {
		return cs.(*kubernetes.Clientset)
	}
	return nil
}
