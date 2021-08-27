package main

import (
	"github.com/hakur/rds-operator/pkg/sidecar"
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	sidecar.RegisterCommand()
	kingpin.Parse()
}
