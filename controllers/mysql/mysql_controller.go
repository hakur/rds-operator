/*
Copyright 2021.

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

package mysql

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/util"
)

// MysqlReconciler reconciles a Mysql object
type MysqlReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=service,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=pod,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=configMap,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Mysql object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (t *MysqlReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.Mysql{}

	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (t *MysqlReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.Mysql{}).
		Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).
		Complete(t)
}

func (t *MysqlReconciler) checkDeleteOrApply(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	if cr.GetDeletionTimestamp().IsZero() {
		// add finalizer mark to CR,make sure CR clean is done by controller first
		if !util.InArray(cr.Finalizers, mysqlFinalizer) {
			cr.ObjectMeta.Finalizers = append(cr.ObjectMeta.Finalizers, mysqlFinalizer)
			if err := t.Update(ctx, cr); err != nil {
				return err
			}
		}
		// apply create CR sub resources
		return t.apply(ctx, cr)
	} else {
		// if finalizer mark exists, that means delete has been failed, try agin
		if util.InArray(cr.Finalizers, mysqlFinalizer) {
			if err := t.clean(ctx, cr); err != nil {
				return err
			}
		}
		// remove finalizer mark, tell k8s I have cleaned all sub resources
		cr.ObjectMeta.Finalizers = util.DelArryElement(cr.ObjectMeta.Finalizers, mysqlFinalizer)
		if err := t.Update(ctx, cr); err != nil {
			return err
		}
	}
	return nil
}

func (t *MysqlReconciler) apply(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	var oldStatefulset appsv1.StatefulSet
	var oldService corev1.Service
	var oldConfigMap corev1.ConfigMap

	statefulset, err := buildMysqlSts(cr)
	if err != nil {
		return err
	}
	service := buildMysqlService(cr)
	containerServices := buildMysqlContainerServices(cr)
	configMap := buildMyCnfCM(cr)
	for _, containerService := range containerServices {
		var oldContainerService corev1.Service
		if err := t.Get(ctx, client.ObjectKeyFromObject(containerService), &oldContainerService); err != nil {
			if err := client.IgnoreNotFound(err); err == nil {
				// add finalizer mark to CR,make sure CR clean is done by controller first
				if err := ctrl.SetControllerReference(cr, configMap, t.Scheme); err != nil {
					return err
				}
				// if service not exists, create it now
				if err := t.Create(ctx, containerService); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			containerService.ResourceVersion = oldContainerService.ResourceVersion
			containerService.Spec.ClusterIP = oldContainerService.Spec.ClusterIP
			// if service exists, update it now
			if err := t.Update(ctx, containerService); err != nil {
				return err
			}
		}
	}

	if err := t.Get(ctx, client.ObjectKeyFromObject(configMap), &oldConfigMap); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, configMap, t.Scheme); err != nil {
				return err
			}
			// if configMap not exists, create it now
			if err := t.Create(ctx, configMap); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if configMap exists, update it now
		if err := t.Update(ctx, configMap); err != nil {
			return err
		}
	}

	if err := t.Get(ctx, client.ObjectKeyFromObject(statefulset), &oldStatefulset); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, statefulset, t.Scheme); err != nil {
				return err
			}
			// if deployment not exist, create it
			if err := t.Create(ctx, statefulset); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		if err := t.Update(ctx, statefulset); err != nil {
			return err
		}
	}

	if err := t.Get(ctx, client.ObjectKeyFromObject(service), &oldService); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, service, t.Scheme); err != nil {
				return err
			}
			//if service not exists, create it
			if err := t.Create(ctx, service); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service exists, update it
		service.ResourceVersion = oldService.ResourceVersion
		service.Spec.ClusterIP = oldService.Spec.ClusterIP
		if err := t.Update(ctx, service); err != nil {
			return err
		}
	}
	return nil
}

func (t *MysqlReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	var statefulsets appsv1.StatefulSetList
	var services corev1.ServiceList
	var configMaps corev1.ConfigMapList
	if err = t.List(ctx, &statefulsets, client.InNamespace(cr.Namespace), client.MatchingLabels(buildMysqlLabels(cr))); err == nil {
		for _, statefulset := range statefulsets.Items {
			if err = t.Delete(ctx, &statefulset); err != nil {
				return err
			}
		}
	}

	if err = t.List(ctx, &services, client.InNamespace(cr.Namespace), client.MatchingLabels(buildMysqlLabels(cr))); err == nil {
		for _, service := range services.Items {
			if err = t.Delete(ctx, &service); err != nil {
				return err
			}
		}
	}

	if err = t.List(ctx, &configMaps, client.InNamespace(cr.Namespace), client.MatchingLabels(buildMysqlLabels(cr))); err == nil {
		for _, configMap := range configMaps.Items {
			if err = t.Delete(ctx, &configMap); err != nil {
				return err
			}
		}
	}

	return nil
}
