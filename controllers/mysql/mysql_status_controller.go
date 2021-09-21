package mysql

import (
	"context"
	"strconv"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
)

func GetMysqlHosts(cr *rdsv1alpha1.Mysql) (hosts []string) {
	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		hosts = append(hosts, cr.Name+"-mysql-"+strconv.Itoa(i))
	}
	return
}

func GetMysqlDataSources(cr *rdsv1alpha1.Mysql) (ds []*mysql.DSN) {
	for _, v := range GetMysqlHosts(cr) {
		ds = append(ds, &mysql.DSN{
			Host:     v + "." + cr.Namespace,
			Port:     3306,
			Username: "root",
			Password: *cr.Spec.RootPassword,
			DBName:   "mysql",
		})
	}
	return
}

// checkClusterStatus check cluster if is running , if not running, try to boostrap cluster
func (t *MysqlReconciler) checkClusterStatus(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	var clusterManager mysql.ClusterManager
	var dataSources = GetMysqlDataSources(cr)

	// set default values
	cr.Status.Members = GetMysqlHosts(cr)
	cr.Status.HealthyMembers = []string{}
	cr.Status.Phase = rdsv1alpha1.MysqlPhaseNotReady

	switch cr.Spec.ClusterMode {
	case rdsv1alpha1.ModeMGRSP:
		clusterManager = &mysql.MGRSP{DataSrouces: dataSources}
	case rdsv1alpha1.ModeSemiSync:
		clusterManager = &mysql.SemiSync{DataSrouces: dataSources}
	}

	if err = clusterManager.StartCluster(ctx); err != nil {
		return err
	}

	for _, v := range clusterManager.HealthyMembers(ctx) {
		cr.Status.HealthyMembers = append(cr.Status.HealthyMembers, v.Host)
	}

	if len(cr.Status.HealthyMembers) == len(cr.Status.Members) {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseRunning
	} else {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseNotReady
	}

	return err
}
