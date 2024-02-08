PLATFORMS := windows/386 windows/amd64 darwin/amd64 darwin/arm64 linux/arm64 linux/amd64 linux/386

# Variables will be evaluated on use > this will also work within a loop
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

# Your platform
build:
	go build -o bin/a9s main.go
	go build -o bin/kubectl-a9s main.go


# All platforms
build_all: $(PLATFORMS)

# Looping through all platforms
$(PLATFORMS):
	$(info Building $@...)	
	GOOS=$(os) GOARCH=$(arch) go build -o bin/a9s-$(os)-$(arch) main.go

test:
	go test ./...

# Further reading: https://vic.demuzere.be/articles/golang-makefile-crosscompile/