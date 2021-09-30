package util

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecPodOnceOpts ExecPodOnce exec a command in a pod options
type ExecPodOnceOpts struct {
	RestConfig    *rest.Config
	KubeClient    *kubernetes.Clientset
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string
	Args          []string
	Timeout       time.Duration
}

// ExecPodOnce exec a command in a pod
func ExecPodOnce(opts ExecPodOnceOpts) (result []byte, err error) {
	restclient := opts.KubeClient.CoreV1().RESTClient().Post()
	req := restclient.Resource("pods").Name(opts.PodName).Namespace(opts.Namespace).SubResource("exec").Timeout(opts.Timeout)

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.ContainerName,
		Command:   opts.Command,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(opts.RestConfig, "POST", req.URL())
	if err != nil {
		err = fmt.Errorf("spdy client exec command error -> %s", err.Error())
		return
	}

	script := strings.Join(opts.Args, "\n")
	stdoutBuf := bytes.NewBuffer([]byte{})
	stderrBuf := bytes.NewBuffer([]byte{})

	err = executor.Stream(remotecommand.StreamOptions{
		Stdout: stdoutBuf,
		Stdin:  strings.NewReader(script),
		Stderr: stderrBuf,
		Tty:    false,
	})

	if err != nil {
		err = fmt.Errorf("spdy executor exec [namespace=%s] [pod=%s] [container=%s] command error -> %s", opts.Namespace, opts.PodName, opts.ContainerName, err.Error())
		return
	}

	result = bytes.TrimSpace(stdoutBuf.Bytes())
	stdErr := bytes.TrimSpace(stderrBuf.Bytes())
	if len(stdErr) > 0 {
		err = fmt.Errorf("spdy executor exec [namespace=%s] [pod=%s] [container=%s] strandard error -> %s", opts.Namespace, opts.PodName, opts.ContainerName, string(stdErr))
	}

	return
}
