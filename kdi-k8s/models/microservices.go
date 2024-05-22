package models

// This file describes the microservice model that will be returned when a deployment is created in the cluster
// It will be used to display the deployment information in the frontend

const (
	RollingUpdateStrategy = "RollingUpdate"
	RecreateStrategy      = "Recreate"
	ABTestingStrategy     = "ab-testing"
	CanaryStrategy        = "canary"
	BlueGreenStrategy     = "blue-green"
)

type Conditions struct {
	Type    string
	Message string
	Reason  string
}

type Container struct {
	Name  string
	Image string
	Port  int32
}

// Microservice represents a deployed microservice
type Microservice struct {
	Name       string
	Namespace  string
	Replicas   int32
	Labels     map[string]string
	Selectors  map[string]string
	Strategy   string // The deployment strategy used
	Conditions []Conditions
	Containers []Container
}
