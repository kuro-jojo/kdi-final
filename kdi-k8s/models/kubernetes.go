package models

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeObjectSet struct {
	Deployment            *appsv1.Deployment
	Service               *corev1.Service
	Namespace             *corev1.Namespace
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
	SetNamespace(namespace string)
	Get(ctx context.Context, name string, opts metav1.GetOptions) error
	Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) error
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
