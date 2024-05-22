package models

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

func (s *Service) SetNamespace(namespace string) {
	s.Service.Namespace = namespace
}

func (s *Service) Get(ctx context.Context, name string, opts metav1.GetOptions) error {
	_s, err := s.Clientset.CoreV1().Services(s.GetNamespace()).Get(ctx, name, opts)
	if err != nil {
		return err
	}
	s.Service = _s
	return nil
}

func (s *Service) Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) error {
	service, ok := obj.(*Service)
	if !ok {
		return fmt.Errorf("invalid type for Service object")
	}
	createdDep, err := s.Clientset.CoreV1().Services(s.Service.Namespace).Create(ctx, service.Service, opts)
	if err != nil {
		return err
	}
	s.Service = createdDep
	return nil
}
