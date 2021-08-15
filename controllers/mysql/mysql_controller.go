package mysql

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
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
//+kubebuilder:rbac:groups=v1,resources=secret,verbs=get;list;watch;create;update;delete
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
		// Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&corev1.Secret{}). // Warning!!! do not open this line, will cause controller Reconcile loop,then get stuck in memory leak, beacause I have called reconciler.Apllyxxxx functions, this Apllyxxxx fucntions already called ctrl.SetControllerReference， so there is no require for manual gc. but gc interface in this controller "clean" function, if you want do some extra gc
		// Watches(&source.Kind{Type: &corev1.PersistentVolumeClaim{}},
		// 	&reconciler.PVCCleaner{}).
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
	if err = t.applyMysql(ctx, cr); err != nil {
		return err
	}

	if err = t.applyProxySQL(ctx, cr); err != nil {
		return err
	}

	if err = reconciler.RemovePVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}

// applyProxySQL create or update proxySQL resources
func (t *MysqlReconciler) applyProxySQL(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	secret := buildSecret(cr)
	service := buildProxySQLService(cr)
	statefulset, err := buildProxySQLSts(cr)
	if err != nil {
		return err
	}

	if err := reconciler.ApplyService(t.Client, ctx, service, cr, t.Scheme); err != nil {
		return err
	}

	if err := reconciler.ApplySecret(t.Client, ctx, secret, cr, t.Scheme); err != nil {
		return err
	}

	if err := reconciler.ApplyStatefulSet(t.Client, ctx, statefulset, cr, t.Scheme); err != nil {
		return err
	}

	return nil
}

// applyMysql create or update mysql resources
func (t *MysqlReconciler) applyMysql(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	statefulset, err := buildMysqlSts(cr)
	if err != nil {
		return err
	}
	service := buildMysqlService(cr)
	containerServices := buildMysqlContainerServices(cr)
	configMap := buildMyCnfCM(cr)
	secret := buildSecret(cr)

	if err = reconciler.ApplySecret(t.Client, ctx, secret, cr, t.Scheme); err != nil {
		return err
	}

	for _, containerService := range containerServices {
		if err = reconciler.ApplyService(t.Client, ctx, containerService, cr, t.Scheme); err != nil {
			return err
		}
	}

	if err = reconciler.ApplyConfigMap(t.Client, ctx, configMap, cr, t.Scheme); err != nil {
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

// clean remove unreferenced sub resources, such as mark pvc delete date
func (t *MysqlReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	if err = reconciler.AddPVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}
	return nil
}
