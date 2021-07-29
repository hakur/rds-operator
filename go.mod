module github.com/hakur/rds-operator

go 1.16

require (
	github.com/bombsimon/logrusr v1.1.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/jinzhu/copier v0.3.2
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gorm.io/driver/mysql v1.1.1 // indirect
	gorm.io/gorm v1.21.12 // indirect
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
