/*
Copyright 2026.

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

	vectorv1alpha1 "qdrant.cloud/qdrant-operator/api/v1alpha1" // Ensure this matches your actual module path
)

// QdrantClusterReconciler reconciles a QdrantCluster object
type QdrantClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=vector.qdrant.cloud,resources=qdrantclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=vector.qdrant.cloud,resources=qdrantclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vector.qdrant.cloud,resources=qdrantclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QdrantCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.1/pkg/reconcile
func (r *QdrantClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var qdrant vectorv1alpha1.QdrantCluster

	if err := r.Get(ctx, req.NamespacedName, &qdrant); err != nil {
		if client.IgnoreNotFound(err) == nil {
			log.Info("QdrantCluster resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get QdrantCluster")
		return ctrl.Result{}, err
	}

	log.Info("Successfully fetched QdrantCluster", "Name", qdrant.Name, "Version", qdrant.Spec.Version)

	// 2. TODO: Create the StatefulSet
	found, err := r.getQdrantImageFromStatefulSet(ctx, &qdrant)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new StatefulSet for QdrantCluster", "StatefulSet.Namespace", qdrant.Namespace, "StatefulSet.Name", qdrant.Name)
		storageSize := r.statefulSetForQdrant(&qdrant)
		err = r.Create(ctx, storageSize)
		if err != nil {
			log.Error(err, "Failed to create new StatefulSet", "StatefulSet.Namespace", qdrant.Namespace, "StatefulSet.Name", qdrant.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	desiredVersion := qdrant.Spec.Version
	currentImage := found.Spec.Template.Spec.Containers[0].Image
	if r.isImageMismatch(desiredVersion, currentImage) {
		log.Info("Updating StatefulSet with new image version", "StatefulSet.Namespace", qdrant.Namespace, "StatefulSet.Name", qdrant.Name, "CurrentImage", currentImage, "DesiredVersion", desiredVersion)
		found.Spec.Template.Spec.Containers[0].Image = "qdrant/qdrant:" + desiredVersion
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update StatefulSet with new image version", "StatefulSet.Namespace", qdrant.Namespace, "StatefulSet.Name", qdrant.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}
	// 3. TODO: Create the Service

	foundSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: qdrant.Name + "-svc", Namespace: qdrant.Namespace}, foundSvc)

	if err != nil && errors.IsNotFound(err) {
		svc := r.serviceForQdrant(&qdrant)
		log.Info("Creating a new Service for QdrantCluster", "Service.Namespace", qdrant.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", qdrant.Namespace, "Service.Name", svc.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return ctrl.Result{}, err
	}
	// 4. TODO: Update the Status

	return ctrl.Result{}, nil
}

func (*QdrantClusterReconciler) isImageMismatch(desiredVersion string, currentImage string) bool {
	desiredImage := "qdrant/qdrant:" + desiredVersion
	return desiredImage != currentImage
}

func (r *QdrantClusterReconciler) getQdrantImageFromStatefulSet(ctx context.Context, q *vectorv1alpha1.QdrantCluster) (*appsv1.StatefulSet, error) {
	found := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: q.Name, Namespace: q.Namespace}, found)
	if err != nil {
		return nil, err
	}
	return found, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QdrantClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vectorv1alpha1.QdrantCluster{}).
		Named("qdrantcluster").
		Complete(r)
}

func (r *QdrantClusterReconciler) statefulSetForQdrant(q *vectorv1alpha1.QdrantCluster) *appsv1.StatefulSet {
	labels := map[string]string{"app": "qdrant", "qdrant_cluster": q.Name}
	replicas := q.Spec.Size

	storageSize := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      q.Name,
			Namespace: q.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "qdrant/qdrant:" + q.Spec.Version,
						Name:  "qdrant",
						Ports: []corev1.ContainerPort{
							{ContainerPort: q.Spec.HTTPPort, Name: "http"},
							{ContainerPort: q.Spec.GRPCPort, Name: "grpc"},
						},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(q, storageSize, r.Scheme)
	return storageSize
}

func (r *QdrantClusterReconciler) serviceForQdrant(q *vectorv1alpha1.QdrantCluster) *corev1.Service {
	labels := map[string]string{"app": "qdrant", "qdrant_cluster": q.Name}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      q.Name + "-svc",
			Namespace: q.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{Name: "http", Port: q.Spec.HTTPPort, TargetPort: intstr.FromInt(int(q.Spec.HTTPPort))},
				{Name: "grpc", Port: q.Spec.GRPCPort, TargetPort: intstr.FromInt(int(q.Spec.GRPCPort))},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	ctrl.SetControllerReference(q, svc, r.Scheme)
	return svc
}
