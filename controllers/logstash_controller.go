/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vstarv1 "github.com/Gentleelephant/logstash-operator/api/v1"
)

// LogstashReconciler reconciles a Logstash object
type LogstashReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vstar.my.birdhk,resources=logstashes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vstar.my.birdhk,resources=logstashes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vstar.my.birdhk,resources=logstashes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Logstash object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *LogstashReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// TODO(user): your logic here
	// deploy logstash
	var logstash vstarv1.Logstash
	var deployment appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &logstash); err != nil {
		l.Error(err, "unable to fetch Logstash")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	err := r.Get(ctx, req.NamespacedName, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			l.Info("deployment not found")
			deployment = r.CreateDeployment(ctx, &logstash)
			fmt.Printf("deployment: %+v")
			err := r.Create(ctx, &deployment)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}
	//binding deployment to podsbook
	if err = ctrl.SetControllerReference(&logstash, &deployment, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogstashReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vstarv1.Logstash{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func (r *LogstashReconciler) CreateDeployment(ctx context.Context, logstash *vstarv1.Logstash) appsv1.Deployment {
	var deployment = appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      logstash.Name,
			Namespace: logstash.Namespace,
			Labels: map[string]string{
				"app": logstash.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &logstash.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": logstash.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": logstash.Name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            logstash.Name,
							Image:           logstash.Spec.Image,
							Command:         []string{"tail", "-f", "/dev/null"},
							ImagePullPolicy: "Always",
						},
					},
				},
			},
		},
	}
	return deployment
}
