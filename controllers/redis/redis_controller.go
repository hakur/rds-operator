package redis

import (
	"context"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/hakur/rds-operator/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const RedisFinlizer = "redis.rds.hakurei.cn"

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=redis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=redis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=redis/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=service,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=pod,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=configMap,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// RedisReconciler redis.rds.hakurei.cn crd controller
type RedisReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (t *RedisReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.Redis{}).
		//Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&appsv1.Deployment{}). //Warning!!! do not open this line, will cause controller Reconcile loop,then get stuck in memory leak, beacause I have called reconciler.Apllyxxxx functions, this Apllyxxxx fucntions already called ctrl.SetControllerReference， so there is no require for manual gc. but gc interface in this controller "clean" function, if you want do some extra gc
		Complete(t)
}

// Reconcile runtime interface implement, accept CR update add delete events
func (t *RedisReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.Redis{}
	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	return ctrl.Result{}, nil
}

// checkDeleteOrApply check cr should delete or apply
func (t *RedisReconciler) checkDeleteOrApply(ctx context.Context, cr *rdsv1alpha1.Redis) (err error) {
	if cr.GetDeletionTimestamp().IsZero() {
		// add finalizer mark to CR,make sure CR clean is done by controller first
		if !util.InArray(cr.Finalizers, RedisFinlizer) {
			cr.ObjectMeta.Finalizers = append(cr.ObjectMeta.Finalizers, RedisFinlizer)
			if err := t.Update(ctx, cr); err != nil {
				return err
			}
		}
		// add bootstrap process worker
		return t.apply(ctx, cr)
	} else {
		// if finalizer mark exists, that means delete has been failed, try agin
		if util.InArray(cr.Finalizers, RedisFinlizer) {
			if err := t.clean(ctx, cr); err != nil {
				return err
			}
		}
		// remove finalizer mark, tell k8s I have cleaned all sub resources
		cr.ObjectMeta.Finalizers = util.DelArryElement(cr.ObjectMeta.Finalizers, RedisFinlizer)
		if err := t.Update(ctx, cr); err != nil {
			return err
		}
	}
	return nil
}

// addBootstrapWorker add worker process thread to bootstrap redis nodes
func (t *RedisReconciler) apply(ctx context.Context, cr *rdsv1alpha1.Redis) (err error) {
	statefulset, err := buildRedisSts(cr)
	if err != nil {
		return err
	}

	deployment, err := buildProxyDeploy(cr)
	if err != nil {
		return err
	}

	redisService := buildRedisSvc(cr)
	proxyService := buildProxySvc(cr)

	// redis servers
	if err = reconciler.ApplyStatefulSet(t.Client, ctx, statefulset, cr, t.Scheme); err != nil {
		return err
	}

	// redis cluster proxy
	if err = reconciler.ApplyDeployment(t.Client, ctx, deployment, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplyService(t.Client, ctx, redisService, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplyService(t.Client, ctx, proxyService, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.RemovePVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}

// clean remove unreferenced sub resources, such as mark pvc delete date
func (t *RedisReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Redis) (err error) {
	if err = reconciler.AddPVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}
	return nil
}
