package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects.
var SchemeGroupVersion = GroupVersion

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	SchemeBuilder.Register(
		&Mysql{}, &MysqlList{},
		&Redis{}, &RedisList{},
		&MysqlBackup{}, &MysqlBackupList{},
		&ProxySQL{}, &ProxySQLList{},
	)
}
