package mysql

import (
	"context"
	"fmt"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/controllers/mysql/builder"
	"github.com/hakur/rds-operator/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// checkClusterStatus check cluster if is running , if not running, try to boostrap cluster
func (t *MysqlReconciler) checkClusterStatus(ctx context.Context, cr *rdsv1alpha1.Mysql) (err error) {
	if cr.DeletionGracePeriodSeconds != nil {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseTerminating
		return nil
	}

	var pods corev1.PodList
	if err = t.List(ctx, &pods, client.InNamespace(cr.Namespace), client.MatchingLabels(builder.BuildMysqlLabels(cr))); err != nil && client.IgnoreNotFound(err) != nil {
		return err
	}

	cr.Status.Members = GetMysqlHosts(cr)
	if cr.Status.Phase == "" {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseCreating
	}

	var podNames []string
	firstRunningPodName := ""
	allPodsRunning := true
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			firstRunningPodName = pod.Name
			podNames = append(podNames, pod.Name)
			break
		} else {
			allPodsRunning = false
		}
	}

	if allPodsRunning {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseNotReady
	}

	if firstRunningPodName == "" {
		return types.ErrPodNotRunning
	}

	master, err := t.Helper.FindMaster(cr.Namespace, firstRunningPodName)
	if err != nil {
		return
	}

	if err = t.Helper.StartCluster(cr.Namespace, podNames); err != nil {
		return err
	}

	if master == "" {
		err = fmt.Errorf("%w", types.ErrMasterNoutFound)
		return err
	}

	master = strings.Trim(master, ",")

	cr.Status.Masters = strings.Split(master, ",")

	if master != "" {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseRunning
	}

	if cr.DeletionGracePeriodSeconds != nil {
		cr.Status.Phase = rdsv1alpha1.MysqlPhaseTerminating
	}

	return err
}
