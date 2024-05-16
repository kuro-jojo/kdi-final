package models

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KubeObjectSet struct {
	Deployment            *appsv1.Deployment
	Service               *corev1.Service
	ConfigMap             *corev1.ConfigMap
	ServiceAccount        *corev1.ServiceAccount
	Secret                *corev1.Secret
	PersistentVolumeClaim *corev1.PersistentVolumeClaim
	PersistentVolume      *corev1.PersistentVolume
	Ingress               *networking.Ingress
	Role                  *rbac.Role
	RoleBinding           *rbac.RoleBinding
}

type KubeObject interface {
	GetName() string
	GetNamespace() string
	Get(ctx context.Context, name string, opts metav1.GetOptions) error
	Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) (KubeObject, error)
}

// Deployment

type Deployment struct {
	Clientset  *kubernetes.Clientset
	Deployment *appsv1.Deployment
}

func (d *Deployment) GetName() string {
	return d.Deployment.Name
}

func (d *Deployment) GetNamespace() string {
	return d.Deployment.Namespace
}

func (d *Deployment) Get(ctx context.Context, name string, opts metav1.GetOptions) error {
	_, err := d.Clientset.AppsV1().Deployments(d.GetNamespace()).Get(ctx, name, opts)
	return err
}

func (d *Deployment) Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) (KubeObject, error) {
	dep, ok := obj.(*Deployment)
	if !ok {
		return nil, fmt.Errorf("invalid type for deployment object")
	}
	createdDep, err := d.Clientset.AppsV1().Deployments(d.Deployment.Namespace).Create(ctx, dep.Deployment, opts)
	if err != nil {
		return nil, err
	}
	return &Deployment{Clientset: d.Clientset, Deployment: createdDep}, nil
}

// Service
type Service struct {
	Clientset *kubernetes.Clientset
	Service   *corev1.Service
}

func (s *Service) GetName() string {
	return s.Service.Name
}

func (s *Service) GetNamespace() string {
	return s.Service.Namespace
}

func (s *Service) Get(ctx context.Context, name string, opts metav1.GetOptions) error {
	_, err := s.Clientset.CoreV1().Services(s.GetNamespace()).Get(ctx, name, opts)
	return err
}

func (s *Service) Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) (KubeObject, error) {
	service, ok := obj.(*Service)
	if !ok {
		return nil, fmt.Errorf("invalid type for Service object")
	}
	createdDep, err := s.Clientset.CoreV1().Services(s.Service.Namespace).Create(ctx, service.Service, opts)
	if err != nil {
		return nil, err
	}
	return &Service{Clientset: s.Clientset, Service: createdDep}, nil
}

type DeploymentInfo struct {
	DockerImage string `json:"dockerImage"`
	ServicePort int    `json:"servicePort"`
	ServiceType string `json:"serviceType"`
	Namespace   string `json:"namespace"`
	ChartName   string `json:"url"`
	ReleaseName string `json:"releaseName"`
}

type RepoEntry struct {
	RepoUrl     string `json:"repoUrl"`
	RepoName    string `json:"repoName"`
	ChartName   string `json:"chart"`
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releaseName"`
}
