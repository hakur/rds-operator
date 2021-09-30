package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	RegisterCommand()
	kingpin.Parse()
}
