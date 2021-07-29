package redis

import (
	"context"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	return
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
	var oldStatefulset appsv1.StatefulSet
	var oldDeployment appsv1.Deployment
	var oldService corev1.Service

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

	// redis cluster proxy
	if err := t.Get(ctx, client.ObjectKeyFromObject(deployment), &oldDeployment); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, deployment, t.Scheme); err != nil {
				return err
			}
			// if deployment not exist, create it
			if err := t.Create(ctx, deployment); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		if err := t.Update(ctx, deployment); err != nil {
			return err
		}
	}

	if err := t.Get(ctx, client.ObjectKeyFromObject(redisService), &oldService); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, redisService, t.Scheme); err != nil {
				return err
			}
			//if service not exists, create it
			if err := t.Create(ctx, redisService); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service exists, update it
		redisService.ResourceVersion = oldService.ResourceVersion
		redisService.Spec.ClusterIP = oldService.Spec.ClusterIP
		if err := t.Update(ctx, redisService); err != nil {
			return err
		}
	}

	if err := t.Get(ctx, client.ObjectKeyFromObject(proxyService), &oldService); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// add finalizer mark to CR,make sure CR clean is done by controller first
			if err := ctrl.SetControllerReference(cr, proxyService, t.Scheme); err != nil {
				return err
			}
			//if service not exists, create it
			if err := t.Create(ctx, proxyService); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service exists, update it
		proxyService.ResourceVersion = oldService.ResourceVersion
		proxyService.Spec.ClusterIP = oldService.Spec.ClusterIP
		if err := t.Update(ctx, proxyService); err != nil {
			return err
		}
	}
	return
}

// clean remove related resources
func (t *RedisReconciler) clean(ctx context.Context, cr *rdsv1alpha1.Redis) (err error) {
	var statefulsets appsv1.StatefulSetList
	if err = t.List(ctx, &statefulsets, client.InNamespace(cr.Namespace), client.MatchingLabels(buildRedisLabels(cr))); err == nil {
		for _, statefulset := range statefulsets.Items {
			if err = t.Delete(ctx, &statefulset); err != nil {
				return err
			}
		}
	}

	var deployments appsv1.DeploymentList
	if err = t.List(ctx, &deployments, client.InNamespace(cr.Namespace), client.MatchingLabels(buildProxyLabels(cr))); err == nil {
		for _, deployment := range deployments.Items {
			if err = t.Delete(ctx, &deployment); err != nil {
				return err
			}
		}
	}

	var services corev1.ServiceList
	if err = t.List(ctx, &services, client.InNamespace(cr.Namespace), client.MatchingLabels(buildProxyLabels(cr))); err == nil {
		for _, service := range services.Items {
			if err = t.Delete(ctx, &service); err != nil {
				return err
			}
		}
	}

	if err = t.List(ctx, &services, client.InNamespace(cr.Namespace), client.MatchingLabels(buildRedisLabels(cr))); err == nil {
		for _, service := range services.Items {
			if err = t.Delete(ctx, &service); err != nil {
				return err
			}
		}
	}

	return nil
}
