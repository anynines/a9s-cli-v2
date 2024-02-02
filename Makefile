BINARY=a9s
PLATFORMS=darwin linux windows
ARCHITECTURES=386 amd64
# the cross product doesnt work as there are combinations that are unwanted or even unsupported
# darwin/386 for example does not work.

# Your platform
build:
	go build -o bin/a9s main.go
	go build -o bin/kubectl-a9s main.go

# All platforms
build_all:
  	# Cross product darwin/386, darwin/amd64, linux/386, ...
	$(foreach GOOS, $(PLATFORMS),\
		$(foreach GOARCH, $(ARCHITECTURES),\
			$(info Building for platform/arch: $(GOOS)/$(GOARCH)) ; $(newline) \
			GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/$(BINARY)-$(GOOS)-$(GOARCH) main.go $(newline) \
			))
# GOOS=windows GOARCH=amd64 go build -o bin/a9s-windows-amd64 main.go
# GOOS=windows GOARCH=amd64 go build -o bin/a9s-windows-amd64 main.go


test:
	go test ./...
