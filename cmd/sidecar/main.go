package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	new(MysqlCommand).Register()
	kingpin.Parse()
}
