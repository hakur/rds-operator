package types

import "errors"

var (
	ErrMasterNoutFound                 = errors.New("master not found")
	ErrPodNotRunning                   = errors.New("pod is not running")
	ErrReplicasNotDesired              = errors.New("pod replicas not desired")
	ErrContainerNotFound               = errors.New("container not found")
	ErrCtxTimeout                      = errors.New("context timeout")
	ErrMyqlConnectFaild                = errors.New("mysql connect failed")
	ErrMysqlStartSemiSyncClusterFailed = errors.New("mysql start semi sync cluster with single primary mode failed")
	ErrMysqlStartMGRSPClusterFailed    = errors.New("mysql start group relication cluster with single primary mode failed")
	ErrMysqlSemiSyncIsAlreadyRunning   = errors.New("mysql group relication is already running")
	ErrMysqlMGRIsAlreadyRunning        = errors.New("mysql group relication is already running")
	ErrMysqlFindMasterFromSalveFailed  = errors.New("mysql try to find master from query slave instance failed")
)
