package mysql

import (
	"context"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlboostraps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlboostraps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=mysqlboostraps/finalizers,verbs=update

// MysqlBootstrapReconciler MysqlBootstrap.rds.hakurei.cn crd controller
type MysqlBootstrapReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (t *MysqlBootstrapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.MysqlBootstrap{}).
		Complete(t)
}

// Reconcile runtime interface implement, accept CR update add delete events
func (t *MysqlBootstrapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.MysqlBootstrap{}
	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	return
}

// checkDeleteOrApply check cr should delete or apply
func (t *MysqlBootstrapReconciler) checkDeleteOrApply(ctx context.Context, cr *rdsv1alpha1.MysqlBootstrap) (err error) {
	if cr.GetDeletionTimestamp().IsZero() {
		// add finalizer mark to CR,make sure CR clean is done by controller first
		if !util.InArray(cr.Finalizers, mysqlFinalizer) {
			cr.ObjectMeta.Finalizers = append(cr.ObjectMeta.Finalizers, mysqlFinalizer)
			if err := t.Update(ctx, cr); err != nil {
				return err
			}
		}
		// add bootstrap process worker
		return t.addBootstrapWorker(ctx, cr)
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

// addBootstrapWorker add worker process thread to bootstrap mysql nodes
func (t *MysqlBootstrapReconciler) addBootstrapWorker(ctx context.Context, cr *rdsv1alpha1.MysqlBootstrap) (err error) {
	return
}

// clean remove related resources
func (t *MysqlBootstrapReconciler) clean(ctx context.Context, cr *rdsv1alpha1.MysqlBootstrap) (err error) {
	return nil
}
