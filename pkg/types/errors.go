package types

import "errors"

var (
	ErrMasterNoutFound    = errors.New("master not found")
	ErrPodNotRunning      = errors.New("pod is not running")
	ErrReplicasNotDesired = errors.New("pod replicas not desired")
	ErrContainerNotFound  = errors.New("container not found")
)
