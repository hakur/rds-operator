package redis

import (
	"context"
	"fmt"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/hakur/rds-operator/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const RedisFinlizer = "redis.rds.hakurei.cn/v1alpha1"

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
		Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&appsv1.Deployment{}).
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

	redisService := buildRedisSvc(cr)
	secret := buildSecret(cr)

	// redis servers
	if err = reconciler.ApplyStatefulSet(t.Client, ctx, statefulset, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplyService(t.Client, ctx, redisService, cr, t.Scheme); err != nil {
		return err
	}

	if err = reconciler.ApplySecret(t.Client, ctx, secret, cr, t.Scheme); err != nil {
		return err
	}

	if cr.Spec.RedisClusterProxy != nil {
		redisClusterPorxydeployment, err := buildProxyDeploy(cr)
		if err != nil {
			return err
		}

		if err = reconciler.ApplyDeployment(t.Client, ctx, redisClusterPorxydeployment, cr, t.Scheme); err != nil {
			return err
		}

		redisClusterProxyService := buildProxySvc(cr)
		if err = reconciler.ApplyService(t.Client, ctx, redisClusterProxyService, cr, t.Scheme); err != nil {
			return err
		}
	}

	if cr.Spec.Predixy != nil {
		predixydeployment, err := buildPredixyDeploy(cr)
		if err != nil {
			return err
		}

		if err = reconciler.ApplyDeployment(t.Client, ctx, predixydeployment, cr, t.Scheme); err != nil {
			return err
		}

		predixyService := buildPredixySvc(cr)
		if err = reconciler.ApplyService(t.Client, ctx, predixyService, cr, t.Scheme); err != nil {
			return err
		}
	}

	if err = reconciler.RemovePVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}

// clean remove sub resources
func (t *RedisReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Redis) (err error) {
	//clean redis sub resources
	var redisStatefulSets appsv1.StatefulSetList
	if err = t.List(ctx, &redisStatefulSets, client.InNamespace(cr.Namespace), client.MatchingLabels(buildRedisLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range redisStatefulSets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var redisServices corev1.ServiceList
	if err = t.List(ctx, &redisServices, client.InNamespace(cr.Namespace), client.MatchingLabels(buildRedisLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range redisServices.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// clean redis cluster proxy sub resources
	var redisClusterProxyDeployments appsv1.DeploymentList
	if err = t.List(ctx, &redisClusterProxyDeployments, client.InNamespace(cr.Namespace), client.MatchingLabels(buildProxyLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range redisClusterProxyDeployments.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var redisClusterProxyServices corev1.ServiceList
	if err = t.List(ctx, &redisClusterProxyServices, client.InNamespace(cr.Namespace), client.MatchingLabels(buildProxyLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range redisClusterProxyServices.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// clean predixy sub resources
	var predixyDeployments appsv1.DeploymentList
	if err = t.List(ctx, &predixyDeployments, client.InNamespace(cr.Namespace), client.MatchingLabels(buildPredixyLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range predixyDeployments.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var predixyServices corev1.ServiceList
	if err = t.List(ctx, &predixyServices, client.InNamespace(cr.Namespace), client.MatchingLabels(buildPredixyLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range predixyServices.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	// remove cr all pods secret resources
	var secrets corev1.SecretList
	if err = t.List(ctx, &secrets, client.InNamespace(cr.Namespace), client.MatchingLabels(buildRedisLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range secrets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
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
