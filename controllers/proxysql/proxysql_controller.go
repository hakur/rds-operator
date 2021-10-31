package proxysql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/controllers/proxysql/builder"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/hakur/rds-operator/pkg/types"
	"github.com/hakur/rds-operator/util"
	hutil "github.com/hakur/util"
)

const (
	// Finalizer proxysql CR delete mark
	Finalizer = "proxysql.rds.hakurei.cn/v1alpha1"
)

// ProxySQLReconciler reconciles a ProxySQL object
type ProxySQLReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RestConfig *rest.Config
}

//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=proxysqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=proxysqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rds.hakurei.cn,resources=proxysqls/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=service,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete;post
//+kubebuilder:rbac:groups="",resources=pods/logs,verbs=get;post;create;list
//+kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;post;create
//+kubebuilder:rbac:groups=v1,resources=configMap,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=v1,resources=secret,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (t *ProxySQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (r ctrl.Result, err error) {
	cr := &rdsv1alpha1.ProxySQL{}

	if err = t.Get(ctx, req.NamespacedName, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	if err = t.checkDeleteOrApply(ctx, cr); err != nil {
		return r, client.IgnoreNotFound(err)
	}

	// write data to proxysql server
	var proxysqlPods corev1.PodList
	if err = t.List(ctx, &proxysqlPods, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildProxySQLLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, pod := range proxysqlPods.Items {
			if pod.Status.Phase != "Running" {
				continue
			}
			dsn := mysql.DSN{
				Host:     pod.Name + "." + cr.Name + "-proxysql." + cr.Namespace + ".svc",
				Port:     6032,
				Username: cr.Spec.ClusterUser.Username,
				Password: hutil.Base64Decode(cr.Spec.ClusterUser.Password),
			}
			if err = t.syncProxySQLData(ctx, cr, dsn); err != nil {
				return r, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (t *ProxySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdsv1alpha1.ProxySQL{}).
		Owns(&corev1.Service{}).Owns(&appsv1.StatefulSet{}).Owns(&corev1.ConfigMap{}).Owns(&corev1.Secret{}).
		Complete(t)
}

func (t *ProxySQLReconciler) checkDeleteOrApply(ctx context.Context, cr *rdsv1alpha1.ProxySQL) (err error) {
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

func (t *ProxySQLReconciler) apply(ctx context.Context, cr *rdsv1alpha1.ProxySQL) (err error) {
	if err = t.applyProxySQL(ctx, cr); err != nil {
		return err
	}

	if err = reconciler.RemovePVCRetentionMark(t.Client, ctx, cr.Namespace, reconciler.BuildCRPVCLabels(cr, cr)); err != nil {
		return err
	}

	return nil
}

// applyProxySQL create or update proxySQL resources
func (t *ProxySQLReconciler) applyProxySQL(ctx context.Context, cr *rdsv1alpha1.ProxySQL) (err error) {
	proxysqlBuilder := builder.ProxySQLBuilder{CR: cr}
	service := proxysqlBuilder.BuildService()
	statefulset, err := proxysqlBuilder.BuildSts()
	if err != nil {
		return err
	}

	if err := reconciler.ApplyService(t.Client, ctx, service, cr, t.Scheme); err != nil {
		return err
	}

	if err := reconciler.ApplyStatefulSet(t.Client, ctx, statefulset, cr, t.Scheme); err != nil {
		return err
	}

	return nil
}

// clean unreferenced sub resources
func (t *ProxySQLReconciler) clean(ctx context.Context, cr *rdsv1alpha1.ProxySQL) (err error) {

	// clean proxysql sub resources
	var proxySQLStatefulSets appsv1.StatefulSetList
	if err = t.List(ctx, &proxySQLStatefulSets, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildProxySQLLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range proxySQLStatefulSets.Items {
			if err = t.Delete(ctx, &v); err != nil && client.IgnoreNotFound(err) != nil {
				return fmt.Errorf("delete resource in [namespace=%s] [api=%s] [kind=%s] [name=%s] failed -> %s", v.Namespace, v.APIVersion, v.Kind, v.Name, err.Error())
			}
		}
	} else {
		return fmt.Errorf("delete sub resource failed,[namespace=%s] [api=%s] [kind=%s] [cr=%s] , err is -> %s", cr.Namespace, cr.APIVersion, cr.Kind, cr.Name, err.Error())
	}

	var proxyServices corev1.ServiceList
	if err = t.List(ctx, &proxyServices, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildProxySQLLabels(cr))); err == nil && client.IgnoreNotFound(err) == nil {
		for _, v := range proxyServices.Items {
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

func (t *ProxySQLReconciler) syncProxySQLData(ctx context.Context, cr *rdsv1alpha1.ProxySQL, dsn mysql.DSN) (err error) {
	pa, err := mysql.NewProxySQLAdmin(dsn)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		if _, err := pa.Conn.Exec("ROLLBACK"); err != nil {
			return err
		}
		return types.ErrCtxTimeout
	default:
		var deleteMysqlUsers []*mysql.TableMysqlUsers
		var addMysqlUsers []*mysql.TableMysqlUsers
		var deleteMysqlServers []*mysql.TableMysqlServers
		var addMysqlServers []*mysql.TableMysqlServers
		var deleteProxySQLServers []*mysql.TableProxySQLServers
		var addProxySQLServers []*mysql.TableProxySQLServers
		var replicas = 1

		if cr.Spec.Replicas != nil {
			replicas = int(*cr.Spec.Replicas)
		}

		for _, s := range cr.Spec.Mysqls {
			addMysqlServers = append(addMysqlServers, &mysql.TableMysqlServers{
				Hostname: s.Host,
				Port:     s.Port,
			})
		}

		for _, u := range cr.Spec.FrontendUsers {
			addMysqlUsers = append(addMysqlUsers, &mysql.TableMysqlUsers{
				Username:         u.Username,
				Password:         hutil.Base64Decode(u.Password),
				DefaultSchema:    sql.NullString{String: "mysql"},
				Frontend:         true,
				DefaultHostgroup: u.DefaultHostGroup,
			})
		}

		for _, u := range cr.Spec.BackendUsers {
			addMysqlUsers = append(addMysqlUsers, &mysql.TableMysqlUsers{
				Username:         u.Username,
				Password:         hutil.Base64Decode(u.Password),
				DefaultSchema:    sql.NullString{String: "mysql"},
				Backend:          true,
				DefaultHostgroup: u.DefaultHostGroup,
			})
		}

		for i := 0; i < replicas; i++ {
			addProxySQLServers = append(addProxySQLServers, &mysql.TableProxySQLServers{
				Hostname: cr.Name + "-" + strconv.Itoa(i) + "." + cr.Name,
			})
		}

		dbMysqlServers, err := pa.GetMysqlServers(ctx)
		if err != nil {
			return err
		}

		for _, a := range dbMysqlServers {
			found := false
			for _, b := range addMysqlServers {
				if b.Hostname == a.Hostname {
					found = true
					break
				}
			}
			if !found {
				deleteMysqlServers = append(deleteMysqlServers, a)
			}
		}

		dbProxysqlServers, err := pa.GetProxySQLServers(ctx)
		if err != nil {
			return err
		}

		for _, a := range dbProxysqlServers {
			found := false
			for _, b := range addProxySQLServers {
				if b.Hostname == a.Hostname {
					found = true
					break
				}
			}
			if !found {
				deleteProxySQLServers = append(deleteProxySQLServers, a)
			}
		}

		dbMysqlUsers, err := pa.GetMysqlUsers(ctx)
		if err != nil {
			return err
		}

		for _, a := range dbMysqlUsers {
			found := false
			for _, b := range addMysqlUsers {
				if b.Username == a.Username && b.Frontend == a.Frontend {
					found = true
					break
				}
			}
			if !found {
				deleteMysqlUsers = append(deleteMysqlUsers, a)
			}
		}

		if err = pa.Begin(ctx); err != nil {
			return err
		}

		if err = pa.AddMysqlServers(ctx, addMysqlServers); err != nil {
			pa.Rollback(ctx)
			return err
		}

		if err = pa.AddProxySQLServers(ctx, addProxySQLServers); err != nil {
			pa.Rollback(ctx)
			return err
		}

		if err = pa.AddMysqlUsers(ctx, addMysqlUsers); err != nil {
			pa.Rollback(ctx)
			return err
		}

		for _, server := range deleteProxySQLServers {
			if err = pa.RemoveProxySQLServer(ctx, server.Hostname); err != nil {
				pa.Rollback(ctx)
				return err
			}
		}

		for _, server := range deleteMysqlServers {
			if err = pa.RemoveMysqlServer(ctx, server.Hostname); err != nil {
				pa.Rollback(ctx)
				return err
			}
		}

		for _, user := range deleteMysqlUsers {
			if err = pa.RemoveMysqlUser(ctx, user.Username, user.Frontend); err != nil {
				pa.Rollback(ctx)
				return err
			}
		}

		if err = pa.LoadMysqlServersToRuntime(ctx); err != nil {
			return err
		}

		if err = pa.LoadMysqlUsersServersToRuntime(ctx); err != nil {
			return err
		}

		if err = pa.LoadProxySQLServersToRuntime(ctx); err != nil {
			return err
		}

		if err = pa.Commit(ctx); err != nil {
			return err
		}
	}

	return err
}
