build:
	go build -o bin/a9s main.go
	go build -o bin/kubectl-a9s main.go

test:
	go test ./...
