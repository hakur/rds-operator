package mysql

import (
	"fmt"
	"strconv"
	"time"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func GetMysqlHosts(cr *rdsv1alpha1.Mysql) (hosts []string) {
	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		hosts = append(hosts, cr.Name+"-mysql-"+strconv.Itoa(i))
	}

	return
}

type MysqlHelper struct {
	RestConfig *rest.Config
	KubeClient *kubernetes.Clientset
}

func (t *MysqlHelper) FindMaster(podNamespace, podName string) (masterHost string, err error) {
	result, err, _ := util.ExecPodOnce(util.ExecPodOnceOpts{
		RestConfig:    t.RestConfig,
		KubeClient:    t.KubeClient,
		ContainerName: "mysql",
		Command:       []string{"bash"},
		PodName:       podName,
		Namespace:     podNamespace,
		Timeout:       time.Second * 1,
		Args: []string{
			`
			#!/bin/bash
			source /scripts/mysql.sh
			FindMasterServer
			exit
			`,
		},
	})

	if err != nil {
		err = fmt.Errorf("FindMaster error -> %s", err.Error())
		return
	}

	masterHost = string(result)

	return
}

func (t *MysqlHelper) StartCluster(namespace string, pods []string) (err error) {
	for _, pod := range pods {
		_, err, _ = util.ExecPodOnce(util.ExecPodOnceOpts{
			RestConfig:    t.RestConfig,
			KubeClient:    t.KubeClient,
			ContainerName: "mysql",
			Command:       []string{"bash"},
			PodName:       pod,
			Namespace:     namespace,
			Timeout:       time.Second * 5,
			Args: []string{
				`
				#!/bin/bash
				source /scripts/mysql.sh
				StartCluster
				exit
			`,
			},
		})
		if err != nil {
			return fmt.Errorf("StartCluster error -> %s", err.Error())
		}
	}
	return nil
}
