package main

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
}
func main() {
	new(MysqlCommand).Register()
	kingpin.Parse()
}
