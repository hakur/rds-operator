package main

import (
	"bufio"
	"bytes"
	"context"
	"os/exec"
	"strconv"
	"time"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type S3Config struct {
	// Bucket bucket name
	Bucket string
	// Endpoint s3 server endpoint, such as 127.0.0.1 without schema and colon
	Endpoint string
	// Accesseky s3 accessKey
	AccessKey string
	// SecretAccessKey s3 secretAccessKey
	SecretAccessKey string
	SSL             bool
	// Path backup file path on s3
	Path string
}

// MysqlBackupCommand mysql backup command executor
// go run . mysql backup --address=yuxing-mysql-0 --address=yuxing-mysql-1 --address=yuxing-mysql-2 --password=123456 --s3-access-key=minioadmin --s3-secret-access-key=minioadmin --s3-endpoint=192.168.1.4:9000
type MysqlBackupCommand struct {
	// Username username used for backup operation, password use env var MYSQL_PWD
	Username string
	// Password password used for backup operation,  password use env var MYSQL_PWD
	Password string
	// Charset backup output sql charset
	Charset       string
	Zlib          bool
	SSLMode       bool
	StructureOnly bool
	DumpCmd       bool
	GlobalVar     *MysqlGlobalFlagValues
	S3            S3Config
	// LockTable lock table when backup, if enabled, mysqlpump switch to single thread mode
	LockTable bool
	// MysqlPump other custom mysqlpump options to override built-in mysqlpump options
	MysqlPump []string
}

func (t *MysqlBackupCommand) Register(cmd *kingpin.CmdClause) {
	cmd.Action(t.Action)
	cmd.Flag("username", "mysql username used for backup operation, password use env var MYSQL_PWD").Default("root").StringVar(&t.Username)
	cmd.Flag("password", "mysql password used for backup operation, password use env var MYSQL_PWD").Default("").StringVar(&t.Password)
	cmd.Flag("charset", "backup output sql charset").Default("utf8").StringVar(&t.Charset)
	cmd.Flag("zlib", "use zlib compress sql file").Default("false").BoolVar(&t.Zlib)
	cmd.Flag("ssl", "use ssl mode connect mysql").Default("false").BoolVar(&t.SSLMode)
	cmd.Flag("structure-only", "only dump table structure without table data dump").Default("false").BoolVar(&t.StructureOnly)
	cmd.Flag("dump-cmd", "print mysql backup command").Default("true").BoolVar(&t.DumpCmd)
	cmd.Flag("s3-bucket", "s3 bucket name").Default("mysql-backup").StringVar(&t.S3.Bucket)
	cmd.Flag("s3-endpoint", "s3 server endpoint, such as 127.0.0.1 without schema and colon").Default("127.0.0.1:9000").StringVar(&t.S3.Endpoint)
	cmd.Flag("s3-access-key", "s3 accessKey").Default("myAccessKey").StringVar(&t.S3.AccessKey)
	cmd.Flag("s3-secret-access-key", "s3 secretAccessKey").Default("mySecretAccessKey").StringVar(&t.S3.SecretAccessKey)
	cmd.Flag("s3-ssl", "s3 ssl connection mode").Default("false").BoolVar(&t.S3.SSL)
	cmd.Flag("s3-path", "backup file path on s3").Default("default-mysql-cluster").StringVar(&t.S3.Path)
	cmd.Flag("lock-table", "lock table when backup, if enabled, mysqlpump switch to single thread mode").Default("false").BoolVar(&t.LockTable)
	cmd.Flag("mysql-pump", "other custom mysqlpump options to override built-in mysqlpump options").StringsVar(&t.MysqlPump)
}

func (t *MysqlBackupCommand) Action(ctx *kingpin.ParseContext) (err error) {
	var clusterManager mysql.ClusterManager

	backupFileName := time.Now().Format("2006-01-02__15_04_05") + ".sql"
	dataSources := AddressesToDSN(t.GlobalVar.Addresses)
	for _, v := range dataSources {
		v.Username = t.Username
		v.Password = t.Password
		v.DBName = "mysql"
	}

	execCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	switch rdsv1alpha1.ClusterMode(t.GlobalVar.Mode) {
	case rdsv1alpha1.ModeMGRSP:
		clusterManager = &mysql.MGRSP{DataSrouces: dataSources}
	case rdsv1alpha1.ModeSemiSync:
		clusterManager = &mysql.SemiSync{DataSrouces: dataSources, DoubleMasterHA: t.GlobalVar.SemiSyncDoubleMasterHA}
	}

	masters, err := clusterManager.FindMaster(execCtx)
	if len(masters) < 1 || masters[0] == nil {
		logrus.Fatal("master not found")
	}

	master := masters[0]
	// list none system databases, only none system databases is needed to backup
	// mysql system databases are [ mysql information_schema performance_schema sys ]

	var cmdArgs = []string{
		"-h" + master.Host,
		"-P" + strconv.Itoa(master.Port),
		"-u" + master.Username,
		"--add-drop-database",
		"--default-character-set=" + t.Charset,
		"--exclude-databases=mysql,sys,information_schema,performance_schema",
		"--skip-watch-progress",
		"--set-gtid-purged=OFF",
	}

	if t.LockTable {
		cmdArgs = append(cmdArgs, "--add-locks")
		cmdArgs = append(cmdArgs, "--default-parallelism=0")
	}

	if t.Zlib {
		cmdArgs = append(cmdArgs, "--compress-output=zlib")
	}

	if t.SSLMode {
		cmdArgs = append(cmdArgs, "--ssl-mode")
	}

	if t.StructureOnly {
		cmdArgs = append(cmdArgs, "--skip-dump-rows")
	}

	cmd := exec.Command("mysqlpump", cmdArgs...)
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+t.Password)

	if t.DumpCmd {
		logrus.Info("exec command:", cmd.String())
	}

	sqlContentBuf := bytes.NewBuffer([]byte{})
	cmd.Stdout = bufio.NewWriter(sqlContentBuf)

	if err = cmd.Run(); err != nil {
		logrus.WithField("err", err.Error()).Fatal("backup failed")
	}

	// upload to s3 server
	minioClient, err := minio.New(t.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(t.S3.AccessKey, t.S3.SecretAccessKey, ""),
		Secure: t.S3.SSL,
	})

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Info("uploading ", t.S3.Path+"/"+backupFileName, " to s3 server ...")
	_, err = minioClient.PutObject(execCtx, t.S3.Bucket, t.S3.Path+"/"+backupFileName, bufio.NewReader(sqlContentBuf), -1, minio.PutObjectOptions{})

	logrus.Info("uploading ", t.S3.Path+"/"+backupFileName, " to s3 server success")

	return err
}
