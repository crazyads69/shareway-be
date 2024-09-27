dev_server:
	nodemon --exec go run main.go --signal SIGTERM

PHONY: start_dev
