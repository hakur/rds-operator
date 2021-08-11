package mysql

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// NewMGRSinglePrimaryBootOpts new mysql mgr single primary mdoel cluster booter options
type NewMGRSinglePrimaryBootOpts struct {
	DB *gorm.DB
}

// NewMGRSinglePrimaryBoot new mysql mgr single primary mdoel cluster booter
func NewMGRSinglePrimaryBoot(opts NewMGRSinglePrimaryBootOpts) (t *MGRSinglePrimaryBoot) {
	t = new(MGRSinglePrimaryBoot)
	t.Opts = opts
	return t
}

// MGRSinglePrimaryBoot mysql mgr single primary mdoel cluster booter
type MGRSinglePrimaryBoot struct {
	Opts NewMGRSinglePrimaryBootOpts
}

// CheckUserUpdate check mysql client user need to be create or update
func (t *MGRSinglePrimaryBoot) CheckUserUpdate(username, password, domain string, privileges []string, privilegesTarget string) (err error) {
	var data = new(MysqlUserTable)
	// check user exists
	if err = t.Opts.DB.Table("mysql.user").Where("user=? and host=?", username, domain).First(data).Error; err != nil || data == nil {
		// user not exists, create it now
		if err = t.Opts.DB.Exec(fmt.Sprintf("CREATE USER IF NOT EXISTS '%s'@'%s' IDENTIFIED BY '%s'", username, domain, password)).Error; err != nil {
			return err
		}
	}

	// refresh privileges
	if err = t.Opts.DB.Exec(fmt.Sprintf("GRANT %s ON %s TO '%s'@'%s'", strings.Join(privileges, ","), privilegesTarget, username, domain)).Error; err != nil {
		return err
	}

	// refresh privileges
	if err = t.Opts.DB.Exec("FLUSH PRIVILEGES").Error; err != nil {
		return err
	}

	return nil
}

// CheckClusterHasAliveNode if there is a alive node, it will be the candicate to bootstrap node
func (t *MGRSinglePrimaryBoot) CheckClusterHasAliveNode(dsns []string) bool {
	fmt.Println("----", dsns)
	for _, dsn := range dsns {
		db, err := NewDB(dsn)
		if err == nil && db != nil {
			if sqlDB, err := db.DB(); err != nil {
				sqlDB.Close()
			}
			return true
		}
	}
	return false
}

// BootCluster bootstrap mysql cluster, must run on boostrap node
func (t *MGRSinglePrimaryBoot) BootCluster() (err error) {
	if err = t.Opts.DB.Exec("SET GLOBAL group_replication_bootstrap_group=ON;").Error; err != nil {
		return err
	}

	if err = t.Opts.DB.Exec("START GROUP_REPLICATION;").Error; err != nil {
		return err
	}

	return t.Opts.DB.Exec("SET GLOBAL group_replication_bootstrap_group=OFF;").Error
}

// JoinCluster join mysql cluster, must run on slave nodes
func (t *MGRSinglePrimaryBoot) JoinCluster(replicationUser, replicationPassword string) (err error) {
	if err = t.Opts.DB.Exec("RESET MASTER;").Error; err != nil {
		return err
	}

	if err = t.Opts.DB.Exec("CHANGE MASTER TO MASTER_USER='" + replicationUser + "',MASTER_PASSWORD='" + replicationPassword + "' FOR CHANNEL 'group_replication_recovery';").Error; err != nil {
		return err
	}

	return t.Opts.DB.Exec("START GROUP_REPLICATION;").Error
}
