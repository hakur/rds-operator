package mysql

import (
	"fmt"
	"testing"
	"time"

	"context"

	hutil "github.com/hakur/util"
)

func TestGetProxySQLServers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	data, err := pa.GetProxySQLServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("proxysql servers data", data)
}

func TestGetMysqlServers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	data, err := pa.GetMysqlServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("mysql servers data", data)
}

func TestGetMysqlUsers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	data, err := pa.GetProxySQLServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("mysql users data", data)
}

func TestAddProxySQLServers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = pa.AddProxySQLServers(ctx, []*TableProxySQLServers{
		{
			Hostname: "yuxing-proxysql-0",
			Port:     6032,
			Weight:   1,
			Comment:  "",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := pa.GetProxySQLServers(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, ps := range data {
		fmt.Println(ps)
	}
}

func TestRemoveProxySQLServers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = pa.RemoveProxySQLServer(ctx, "yuxing-proxysql-0")
	if err != nil {
		t.Fatal(err)
	}

	data, err := pa.GetProxySQLServers(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for _, ps := range data {
		fmt.Println(ps)
	}
}

func TestAddMysqlServers(t *testing.T) {
	pa, err := NewProxySQLAdmin(DSN{
		Host:     "127.0.0.1",
		Port:     6032,
		Username: "admin",
		Password: "replication_password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pa.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = pa.AddMysqlServers(ctx, []*TableMysqlServers{{Hostname: "yuxing-proxysql-0"}, {Hostname: "yuxing-proxysql-1"}, {Hostname: "yuxing-proxysql-2"}})
	if err != nil {
		t.Fatal(err)
	}

	data, err := pa.GetMysqlServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	hutil.LogJSON(data)
}

func TestRemoveMysqlServers(t *testing.T) {

}

func TestAddMysqlUsers(t *testing.T) {

}

func TestRemoveMysqlUsers(t *testing.T) {

}
