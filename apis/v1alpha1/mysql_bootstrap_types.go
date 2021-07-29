package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// MysqlUser mysql user settings
type MysqlUser struct {
	// Username mysql login account name
	Username string `json:"username"`
	// Password mysql login password of this user
	Password string `json:"password"`
	// Privileges mysql grant sql privileges, for example : []stirng{ "SELECT" ,"REPLICATION CLIENT"} or []string{"ALL PRIVILEGES"}
	Privileges []string `json:"privileges"`
	// Domain user login domain , for example : '%'
	Domain string `json:"domain"`
	// DatabaseTarget which database or tables will granted privileges to this user.
	// for example : grant all privileges on *.* to user xxx@'%' indentified by 'xxxxx', in this case, DatabaseTarget value should be '*.*'
	DatabaseTarget string `json:"databaseTarget"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mcb

// MysqlBootstrap mysql cluster bootstrap spec
// this will create a cluster master status watch
// if all nodes is running, start check master status.
// if master not exists for serveral seconds , will rise first pod to be master
// if master down, and new master found, will make first pod to be an slave/foller,and rejoin to cluster
type MysqlBootstrap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MysqlBootstrapSpec   `json:"spec,omitempty"`
	Status            MysqlBootstrapStatus `json:"status,omitempty"`
}

// MysqlBootstrapSpec bootstrap spec
type MysqlBootstrapSpec struct {
	// MysqlUsers a list of users will be created when initialize cluster
	MysqlUsers []MysqlUser `json:"mysqlUsers"`
}

// MysqlBootstrapStatus bootstrap process status
type MysqlBootstrapStatus struct {
}

//+kubebuilder:object:root=true

// MysqlBootstrapList a list of MysqlBootstrap
type MysqlBootstrapList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlBootstrap `json:"items"`
}
