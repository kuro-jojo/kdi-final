package models

type Microservice struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Replicas      int32  `json:"replicas"`
	EnvironmentID string `json:"environment_id"`
	NamespaceID   string `json:"namespace_id"`
	UserID        string `json:"user_id"`
}

func (m *Microservice) Create() (Microservice, error) {
	return Microservice{}, nil
}
