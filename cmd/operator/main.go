package main

import (

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	scm "sigs.k8s.io/controller-runtime/pkg/scheme"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	mysqlcontrollers "github.com/hakur/rds-operator/controllers/mysql"
	mysqlbackups "github.com/hakur/rds-operator/controllers/mysql_backup"
	proxysqlcontrollers "github.com/hakur/rds-operator/controllers/proxysql"
	rediscontrollers "github.com/hakur/rds-operator/controllers/redis"
	"github.com/hakur/rds-operator/util"

	"github.com/bombsimon/logrusr/v2"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	//+kubebuilder:scaffold:imports
)

var (
	scheme               = runtime.NewScheme()
	metricsAddr          = kingpin.Flag("metrics-bind-address", "metrics http listen address").Default(":8080").String()
	probeAddr            = kingpin.Flag("health-probe-bind-address", "http listen address for liveness check and readyness check").Default(":8081").String()
	enableLeaderElection = kingpin.Flag("leader-elect", "is enable multi operators leader election ，only one operator pod work if enabled leader election").Default("false").Bool()
	namespaceFilter      = kingpin.Flag("namespace", "namespace for crd watching,watch all namespaces if value is empty").Default(util.EnvOrDefault("NAMESPACE", "")).String()
	logLevel             = kingpin.Flag("log-level", "log level this application").Default(util.EnvOrDefault("LOG_LEVEL", "info")).String()
	runController        = kingpin.Flag("run-controller", "run specific operator controller").Default("all").Enum("all", "mysql", "mysqlBackup", "proxysql", "redis")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(rdsv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	kingpin.Parse()
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	parsedLevel, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatal("log level=[%s] is invalid", logLevel)
	}

	logrus.SetFormatter(customFormatter)
	logrus.SetLevel(parsedLevel)
	// logrus.SetReportCaller(true) // bad for logrusr

	// let controllers can list/get/delete/update/create
	prometheusOperatorSchema := scm.Builder{GroupVersion: schema.GroupVersion{Group: "monitoring.coreos.com", Version: "v1"}}
	prometheusOperatorSchema.Register(&monitorv1.ServiceMonitor{}, &monitorv1.ServiceMonitorList{}, &monitorv1.PodMonitor{}, &monitorv1.PodMonitorList{})
	prometheusOperatorSchema.AddToScheme(scheme)
}

func main() {
	ctrl.SetLogger(logrusr.New(logrus.StandardLogger()))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     *metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *enableLeaderElection,
		LeaderElectionID:       "rds.hakurei.cn",
		Namespace:              *namespaceFilter,
	})

	if err != nil {
		logrus.WithField("err", err.Error()).Fatal("run operator controller manager failed")
	}

	if *runController == "all" || *runController == "redis" {
		if err = (&rediscontrollers.RedisReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			logrus.WithField("err", err.Error()).WithField("controller", "Redis").Fatal("could not set up redis.rds.hakurei.cn controller with manager")
		}
	}

	if *runController == "all" || *runController == "mysql" {
		if err = (&mysqlcontrollers.MysqlReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}).SetupWithManager(mgr); err != nil {
			logrus.WithField("err", err.Error()).WithField("controller", "Mysql").Fatal("could not set up mysqls.rds.hakurei.cn controller with manager")
		}
	}

	if *runController == "all" || *runController == "proxysql" {
		if err = (&proxysqlcontrollers.ProxySQLReconciler{
			Client:     mgr.GetClient(),
			Scheme:     mgr.GetScheme(),
			RestConfig: mgr.GetConfig(),
		}).SetupWithManager(mgr); err != nil {
			logrus.WithField("err", err.Error()).WithField("controller", "ProxySQL").Fatal("could not set up proxysqls.rds.hakurei.cn controller with manager")
		}
	}

	if *runController == "all" || *runController == "mysqlBackup" {
		if err = (&mysqlbackups.MysqlBackupReconciler{
			Client:     mgr.GetClient(),
			Scheme:     mgr.GetScheme(),
			RestConfig: mgr.GetConfig(),
		}).SetupWithManager(mgr); err != nil {
			logrus.WithField("err", err.Error()).WithField("controller", " MysqlBackup").Fatal("could not set up mysqlbackups.rds.hakurei.cn controller with manager")
		}
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

//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=podmonitors,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;update;delete

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;delete

//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete

//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;delete

//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;delete;post
//+kubebuilder:rbac:groups="",resources=pods/logs,verbs=get;post;create;list
//+kubebuilder:rbac:groups="",resources=pods/exec,verbs=get;post;create

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

//+kubebuilder:rbac:groups=authentication.k8s.io,resources=tokenreviews,verbs=create
//+kubebuilder:rbac:groups=authentication.k8s.io,resources=subjectaccessreviews,verbs=create
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete
