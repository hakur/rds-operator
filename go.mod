module github.com/hakur/rds-operator

go 1.16

require (
	github.com/bombsimon/logrusr v1.1.0
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/jinzhu/copier v0.3.2
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.9.6
)
