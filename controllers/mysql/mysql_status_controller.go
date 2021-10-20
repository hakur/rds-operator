package mysql

import (
	"context"
	"reflect"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/hakur/util"
	"github.com/sirupsen/logrus"
)

func GetMysqlHosts(cr *rdsv1alpha1.Mysql) (hosts []string) {
	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		hosts = append(hosts, cr.Name+"-mysql-"+strconv.Itoa(i))
	}
	return
}

func GetMysqlDataSources(cr *rdsv1alpha1.Mysql) (ds []*mysql.DSN) {
	mysqlPassword := util.Base64Decode(cr.Spec.ClusterUser.Password)
	for _, v := range GetMysqlHosts(cr) {
		ds = append(ds, &mysql.DSN{
			Host:     v + "." + cr.Namespace,
			Port:     3306,
			Username: cr.Spec.ClusterUser.Username,
			Password: string(mysqlPassword),
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
	masterHosts := cr.Status.Masters
	cr.Status.Members = GetMysqlHosts(cr)
	cr.Status.Masters = []string{}
	cr.Status.HealthyMembers = []string{}
	cr.Status.Phase = rdsv1alpha1.MysqlPhaseNotReady

	switch cr.Spec.ClusterMode {
	case rdsv1alpha1.ModeMGRSP:
		clusterManager = &mysql.MGRSP{DataSrouces: dataSources}
	case rdsv1alpha1.ModeSemiSync:
		if cr.Spec.SemiSync != nil {
			clusterManager = &mysql.SemiSync{DataSrouces: dataSources, DoubleMasterHA: cr.Spec.SemiSync.DoubleMasterHA}
		} else {
			clusterManager = &mysql.SemiSync{DataSrouces: dataSources}
		}
	}

	if err = clusterManager.StartCluster(ctx); err != nil {
		return err
	}

	for _, v := range clusterManager.HealthyMembers(ctx) {
		cr.Status.HealthyMembers = append(cr.Status.HealthyMembers, strings.ReplaceAll(v.Host, "."+cr.Namespace, ""))
	}

	if len(cr.Status.HealthyMembers) == len(cr.Status.Members) {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseRunning
	} else {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseNotReady
	}

	if masters, err := clusterManager.FindMaster(ctx); err == nil {
		for _, master := range masters {
			cr.Status.Masters = append(cr.Status.Masters, strings.ReplaceAll(master.Host, "."+cr.Namespace, ""))
		}
	} else {
		return err
	}

	if !reflect.DeepEqual(masterHosts, cr.Status.Masters) {
		// master changed, need to notify mysql proxy middleware
		logrus.Debug("master list changed, need to notify mysql proxy middleware")
	}

	return err
}
