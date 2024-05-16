package models

type Container struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Image          string `json:"image"`
	MicroserviceID string `json:"microservice_id"`
}
