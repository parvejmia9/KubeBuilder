/*
Copyright 2024 parvejmia9.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	readerv1 "github.com/parvejmia9/Kubebuilder/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

// BookstoreReconciler reconciles a Bookstore object
type BookstoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=reader.com,resources=bookstores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=reader.com,resources=bookstores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=reader.com,resources=bookstores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bookstore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile

func isDeploymentChanged(bk *readerv1.Bookstore, curdep *appsv1.Deployment) bool {
	if *bk.Spec.DeploymentSpec.Replicas != 0 && *curdep.Spec.Replicas != *bk.Spec.DeploymentSpec.Replicas {
		*curdep.Spec.Replicas = *bk.Spec.DeploymentSpec.Replicas
		return true
	}
	if bk.Spec.DeploymentSpec.Image != "" && curdep.Spec.Template.Spec.Containers[0].Image != bk.Spec.DeploymentSpec.Image {
		curdep.Spec.Template.Spec.Containers[0].Image = bk.Spec.DeploymentSpec.Image
		return true
	}
	if bk.Spec.DeploymentSpec.Name != "" && curdep.Name != bk.Spec.DeploymentSpec.Name {
		curdep.Name = bk.Spec.DeploymentSpec.Name
		return true
	}
	return false
}

func isServiceChanged(bk *readerv1.Bookstore, cursvc *corev1.Service) bool {
	if bk.Spec.ServiceSpec.Name != "" && bk.Spec.ServiceSpec.Name != cursvc.Name {
		cursvc.Name = bk.Spec.ServiceSpec.Name
		fmt.Println("....name...")
		return true
	}
	if bk.Spec.ServiceSpec.Port != 0 && bk.Spec.ServiceSpec.Port != cursvc.Spec.Ports[0].Port {
		cursvc.Spec.Ports[0].Port = bk.Spec.ServiceSpec.Port
		fmt.Println("....port...")
		return true
	}

	// deployment's port is the target port for service

	if bk.Spec.DeploymentSpec.Port != 0 && intstr.FromInt32(bk.Spec.DeploymentSpec.Port) != cursvc.Spec.Ports[0].TargetPort {
		cursvc.Spec.Ports[0].TargetPort = intstr.FromInt32(bk.Spec.DeploymentSpec.Port)
		fmt.Println("....targetport...")
		return true
	}

	return false
}

func (r *BookstoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fmt.Println("<<<<< In Reconcile Function >>>> ")
	fmt.Println("Name: ", req.NamespacedName.Name, "\nNamespace: ", req.NamespacedName.Namespace)

	// slow down the execution to see the changes clearly
	time.Sleep(time.Second * 1)
	// TODO(user): your logic here
	var bk readerv1.Bookstore
	fmt.Println(bk.Spec.DeploymentSpec.Name)
	if err := r.Get(ctx, req.NamespacedName, &bk); err != nil {
		fmt.Println("Unable to get Bookstore", err)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	fmt.Println("Checking Deployment...")
	// Create or Update Deployment

	var deploy appsv1.Deployment

	depName := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      bk.Spec.DeploymentSpec.Name,
	}
	fmt.Println("dep: ", depName)
	if err := r.Get(ctx, depName, &deploy); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		deploy = *newDeployment(&bk)
		if err := r.Create(ctx, &deploy); err != nil {
			fmt.Println("Unable to create Deployment", err)
			return ctrl.Result{}, err
		}
		fmt.Println("New Deployment created")
	}

	if isDeploymentChanged(&bk, &deploy) {
		fmt.Println("Deployment changed")
		if err := r.Client.Update(ctx, &deploy); err != nil {
			fmt.Println("Unable to update Deployment", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Deployment updated")
	}

	fmt.Println("Checking Service...")
	// create or update service
	var svc corev1.Service
	svcName := types.NamespacedName{
		Namespace: req.NamespacedName.Namespace,
		Name:      bk.Spec.ServiceSpec.Name,
	}
	fmt.Println("svc: ", svcName)
	if err := r.Get(ctx, svcName, &svc); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		svc = *newService(&bk)
		if err := r.Create(ctx, &svc); err != nil {
			fmt.Println("Unable to create Service", err)
			return ctrl.Result{}, err
		}
		fmt.Println("New Service created")
	}
	if isServiceChanged(&bk, &svc) {
		fmt.Println("Service changed")
		if err := r.Client.Update(ctx, &svc); err != nil {
			fmt.Println("Unable to update Service", err)
			return ctrl.Result{}, err
		}
		fmt.Println("Service updated")
	}

	return ctrl.Result{}, nil
}

var (
	deployOwnerKey = ".metadata.controller"
	svcOwnerKey    = ".metadata.controller"
	apiGVStr       = readerv1.GroupVersion.String()
	ourKind        = "Bookstore"
)

// SetupWithManager sets up the controller with the Manager.
func (r *BookstoreReconciler) SetupWithManager(mgr ctrl.Manager) error {

	fmt.Println("...... In Manager ......")

	// for deployment
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, deployOwnerKey, func(object client.Object) []string {
		// Get the Deployment object
		deployment := object.(*appsv1.Deployment)
		// Extract the owner
		owner := metav1.GetControllerOf(deployment)
		if owner == nil {
			return nil
		}
		// Make sure it is a Custom Resource
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		// ...and if so, return it.
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	// for service

	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Service{}, deployOwnerKey, func(object client.Object) []string {
		// get the service
		service := object.(*corev1.Service)
		owner := metav1.GetControllerOf(service)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != ourKind {
			return nil
		}
		return []string{owner.Name}
	}); err != nil {

	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&readerv1.Bookstore{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func newDeployment(bk *readerv1.Bookstore) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bk.Spec.DeploymentSpec.Name,
			Namespace: bk.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bk, readerv1.GroupVersion.WithKind("Bookstore")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: bk.Spec.DeploymentSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "bookstore-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "bookstore-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "bookstore-app",
							Image: bk.Spec.DeploymentSpec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: bk.Spec.DeploymentSpec.Port,
								},
							},
						},
					},
				},
			},
		},
	}
}
func newService(bk *readerv1.Bookstore) *corev1.Service {
	labels := map[string]string{
		"app": "bookstore-app",
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      bk.Spec.ServiceSpec.Name,
			Namespace: bk.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(bk, readerv1.GroupVersion.WithKind("Bookstore")),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:       3200,
					TargetPort: intstr.FromInt32(bk.Spec.DeploymentSpec.Port),
					NodePort:   30002,
				},
			},
		},
	}
}
