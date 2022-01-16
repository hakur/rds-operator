package mysqlbackup

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/hakur/rds-operator/util"
)

const (
	// Finalizer mysqlbackups CR delete mark
	Finalizer = "mysqlbackup.rds.hakurei.cn/v1alpha1"
)

// MysqlBackupReconciler reconciles a MysqlBackup object
type MysqlBackupReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlbackups/finalizers,verbs=update

func (t *MysqlBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.MysqlBackup{}

	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (t *MysqlBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.MysqlBackup{}).
		Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&corev1.Secret{}).
		Complete(t)
}

func (t *MysqlBackupReconciler) checkDeleteOrApply(ctx context.Context, cr *rdsv1alpha1.MysqlBackup) (err error) {
	if cr.GetDeletionTimestamp().IsZero() {
		// add finalizer mark to CR,make sure CR clean is done by controller first
		if !util.InArray(cr.Finalizers, Finalizer) {
			cr.ObjectMeta.Finalizers = append(cr.ObjectMeta.Finalizers, Finalizer)
			if err := t.Update(ctx, cr); err != nil {
				return err
			}
		}
		// apply create CR sub resources
		return t.apply(ctx, cr)
	} else {
		// if finalizer mark exists, that means delete has been failed, try agin
		if util.InArray(cr.Finalizers, Finalizer) {
			if err := t.clean(ctx, cr); err != nil {
				return err
			}
		}
		// remove finalizer mark, tell k8s I have cleaned all sub resources
		cr.ObjectMeta.Finalizers = util.DelArryElement(cr.ObjectMeta.Finalizers, Finalizer)
		if err := t.Update(ctx, cr); err != nil {
			return err
		}
	}
	return nil
}

func (t *MysqlBackupReconciler) apply(ctx context.Context, cr *rdsv1alpha1.MysqlBackup) (err error) {
	builder := CronJobBuilder{CR: cr}

	cronjob, err := builder.BuildCronJob()
	if err != nil {
		return
	}

	secret := BuildSecret(cr)

	if err != nil {
		return
	}

	if err = reconciler.ApplyCronJob(t.Client, ctx, cronjob, cr, t.Scheme); err != nil {
		return
	}

	if err = reconciler.ApplySecret(t.Client, ctx, secret, cr, t.Scheme); err != nil {
		return
	}

	return err
}

// clean unreferenced sub resources
func (t *MysqlBackupReconciler) clean(ctx context.Context, cr *rdsv1alpha1.MysqlBackup) (err error) {
	var cronjobs batchv1.CronJobList
	if err = t.List(ctx, &cronjobs, client.InNamespace(cr.Namespace), client.MatchingLabels(BuildLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range cronjobs.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var secrets corev1.SecretList
	if err = t.List(ctx, &secrets, client.InNamespace(cr.Namespace), client.MatchingLabels(BuildLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range secrets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var configMaps corev1.ConfigMapList
	if err = t.List(ctx, &configMaps, client.InNamespace(cr.Namespace), client.MatchingLabels(BuildLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range configMaps.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	return nil
}
