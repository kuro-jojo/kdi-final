package models

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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

func (d *Deployment) SetNamespace(namespace string) {
	d.Deployment.Namespace = namespace
}

func (d *Deployment) Get(ctx context.Context, name string, opts metav1.GetOptions) error {

	_d, err := d.Clientset.AppsV1().Deployments(d.GetNamespace()).Get(ctx, name, opts)
	if err != nil {
		return err
	}
	d.Deployment = _d
	return nil
}

func (d *Deployment) Create(ctx context.Context, obj KubeObject, opts metav1.CreateOptions) error {
	dep, ok := obj.(*Deployment)
	if !ok {
		return fmt.Errorf("invalid type for deployment object")
	}
	creatdedDep, err := d.Clientset.AppsV1().Deployments(d.Deployment.Namespace).Create(ctx, dep.Deployment, opts)
	if err != nil {
		return err
	}
	d.Deployment = creatdedDep
	return nil
}
