module github.com/hakur/rds-operator

go 1.16

require (
	github.com/Rican7/retry v0.3.0
	github.com/bombsimon/logrusr v1.1.0
	github.com/jinzhu/copier v0.3.2
	github.com/sirupsen/logrus v1.8.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gorm.io/driver/mysql v1.1.1
	gorm.io/gorm v1.21.12
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	sigs.k8s.io/controller-runtime v0.9.6
)
