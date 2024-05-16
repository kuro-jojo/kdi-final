package models

type Namespace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ClusterID string `json:"cluster_id"`
}

func (n *Namespace) Create() (Namespace, error) {
	return Namespace{}, nil
}

func (n *Namespace) Update(newNamespace Namespace) (Namespace, error) {
	return Namespace{}, nil
}

func (n *Namespace) UpdateToken(newToken string) (Namespace, error) {
	return Namespace{}, nil
}

func (n *Namespace) Delete() error {
	return nil
}

// Get retrieves a namespace by its ID
func (n *Namespace) Get(namespaceID string) (Namespace, error) {
	return Namespace{}, nil
}

func (n *Namespace) List() ([]Namespace, error) {
	return []Namespace{}, nil
}

// GetProjects retrieves all projects (application) deployed in the namespace
func (n *Namespace) GetProjects() ([]Project, error) {
	return []Project{}, nil
}
