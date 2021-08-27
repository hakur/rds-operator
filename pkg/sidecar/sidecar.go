package sidecar

func RegisterCommand() {
	new(MysqlCommand).Register()
}
