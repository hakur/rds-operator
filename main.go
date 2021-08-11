package main

import (

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	"github.com/bombsimon/logrusr"
	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	mysqlcontrollers "github.com/hakur/rds-operator/controllers/mysql"
	rediscontrollers "github.com/hakur/rds-operator/controllers/redis"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	//+kubebuilder:scaffold:imports
)

var (
	scheme               = runtime.NewScheme()
	metricsAddr          = kingpin.Arg("metrics-bind-address", "metrics http 监听地址").Default(":8080").String()
	probeAddr            = kingpin.Arg("health-probe-bind-address", "liveness 和 readyness 的 http 监听地址").Default(":8081").String()
	enableLeaderElection = kingpin.Arg("leader-elect", "是否启用选举，如果启用选举将只有一个 operator pod 会工作").Default("false").Bool()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(rdsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	kingpin.Parse()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
}

func main() {
	ctrl.SetLogger(logrusr.NewLogger(logrus.StandardLogger()))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     *metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *enableLeaderElection,
		LeaderElectionID:       "rds.hakurei.cn",
	})

	if err != nil {
		logrus.WithField("err", err.Error()).Fatal("无法启动manager")
	}

	if err = (&rediscontrollers.RedisReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logrus.WithField("err", err.Error()).WithField("controller", "Redis").Fatal("could not set up redis.rds.hakurei.cn controller with manager")
	}

	if err = (&mysqlcontrollers.MysqlReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		logrus.WithField("err", err.Error()).WithField("controller", "Mysql").Fatal("could not set up mysqls.rds.hakurei.cn controller with manager")
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		logrus.WithField("err", err.Error()).Fatal("cound not set up controller manager healthz")
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		logrus.WithField("err", err.Error()).Fatal("could not set up controller manager readyz")
	}

	logrus.Info("start controller manager ...")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logrus.WithField("err", err.Error()).Fatal("start controller manager failed")
	}
}
