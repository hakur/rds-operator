all:
	go build -ldflags "-s -w" -o mysql-config 
	docker build -t rumia/mysql-config:v0.0.1 -f Dockerfile .