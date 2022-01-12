package mysql

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/controllers/mysql/builder"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/hakur/rds-operator/pkg/types"
	"github.com/hakur/rds-operator/util"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

const (
	// mysqlFinalizer mysql CR delete mark
	mysqlFinalizer = "mysql.rds.hakurei.cn/v1alpha1"
)

// MysqlReconciler reconciles a Mysql object
type MysqlReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=service,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete;post
//+kubebuilder:rbac:groups="",resources=pods/logs,verbs=get;post;create;list
//+kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;post;create
//+kubebuilder:rbac:groups=v1,resources=configMap,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=secret,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (t *MysqlReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.Mysql{}

	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if cr.GetDeletionTimestamp().IsZero() {
		// check for cluster status
		remoteCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if err = t.checkClusterStatus(remoteCtx, cr); err != nil {
			r.Requeue = true
			r.RequeueAfter = time.Second * 2
			if errors.Is(err, types.ErrPodNotRunning) || errors.Is(err, types.ErrMasterNoutFound) || errors.Is(err, types.ErrContainerNotFound) || errors.Is(err, types.ErrReplicasNotDesired) {
				return r, nil
			}

			return r, err
		}

		if err = t.Status().Update(remoteCtx, cr); err != nil {
			return r, fmt.Errorf("status update failed -> %w", err)
		}

		if cr.Status.Phase != rdsv1alpha1.MysqlPhaseRunning {
			r.Requeue = true
			r.RequeueAfter = time.Second * 2
			return r, nil
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (t *MysqlReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.Mysql{}).
		Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&corev1.Secret{}).Owns(&corev1.Pod{}).
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

		cr.Status.Phase = rdsv1alpha1.MysqlPhaseTerminating
		// remove finalizer mark, tell k8s I have cleaned all sub resources
		cr.ObjectMeta.Finalizers = util.DelArryElement(cr.ObjectMeta.Finalizers, mysqlFinalizer)
		if err := t.Update(ctx, cr); err != nil {
			return err
		}
	}
	return nil
}

func (t *MysqlReconciler) apply(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	if err = t.applyMysql(ctx, cr); err != nil {
		return err
	}

	if cr.Spec.Monitor != nil {
		if err = t.applyMonitor(ctx, cr); err != nil {
			return err
		}
	}

	if err = reconciler.RemovePVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}

func (t *MysqlReconciler) applyMonitor(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	data := builder.BuildPodMonitor(cr)
	var oldData monitorv1.PodMonitor

	if err := t.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if service monitor not exists, create it now
			if err := t.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service monitor exists, update it now
		data.ResourceVersion = oldData.ResourceVersion
		if err := t.Update(ctx, data); err != nil {
			return err
		}
	}

	return
}

// applyMysql create or update mysql resources
func (t *MysqlReconciler) applyMysql(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	mysqlBuilder := builder.MysqlBuilder{CR: cr}
	statefulset, err := mysqlBuilder.BuildSts()
	if err != nil {
		return err
	}

	service := mysqlBuilder.BuildService(cr)
	containerServices := mysqlBuilder.BuildContainerServices(cr)
	cnfConfigMap := mysqlBuilder.BuildMyCnfCM(cr)
	if err != nil {
		return err
	}

	secret := builder.BuildSecret(cr)

	if err = reconciler.ApplySecret(t.Client, ctx, secret, cr, t.Scheme); err != nil {
		return err
	}

	for _, containerService := range containerServices {
		if err = reconciler.ApplyService(t.Client, ctx, containerService, cr, t.Scheme); err != nil {
			return err
		}
	}

	if err = reconciler.ApplyConfigMap(t.Client, ctx, cnfConfigMap, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplyStatefulSet(t.Client, ctx, statefulset, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplyService(t.Client, ctx, service, cr, t.Scheme); err != nil {
		return err
	}

	return nil
}

// clean unreferenced sub resources
func (t *MysqlReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	// here write manual clean codes due to controller runtime SetOwnerRef is not stable
	// clean mysql sub resources
	var mysqlStatefulSets appsv1.StatefulSetList
	if err = t.List(ctx, &mysqlStatefulSets, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range mysqlStatefulSets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var mysqlServices corev1.ServiceList
	if err = t.List(ctx, &mysqlServices, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range mysqlServices.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// clean common sub resources
	var configMaps corev1.ConfigMapList
	if err = t.List(ctx, &configMaps, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range configMaps.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var secrets corev1.SecretList
	if err = t.List(ctx, &secrets, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range secrets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// clean service monitor
	var monitors monitorv1.PodMonitorList
	if err = t.List(ctx, &monitors, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range monitors.Items {
			if err = t.Delete(ctx, v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// add pvc life deadline annotaion mark
	if err = reconciler.AddPVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}
