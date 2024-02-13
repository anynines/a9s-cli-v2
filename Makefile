PLATFORMS := windows/386 windows/amd64 darwin/amd64 darwin/arm64 linux/arm64 linux/amd64 linux/386

# Variables will be evaluated on use > this will also work within a loop
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

timestamp = $(shell date +%s)
version = "v0.10.0"

# Your platform
build:
	$(info Build time is $(timestamp))
	go build -v -ldflags "-X 'github.com/anynines/a9s-cli-v2/cmd.BuildTimestamp="$(timestamp)"' -X 'github.com/anynines/a9s-cli-v2/cmd.CliVersion=$(version)'" -o bin/a9s main.go
	cp bin/a9s bin/kubectl-a9s

# All platforms
build_all: $(PLATFORMS)

# Looping through all platforms
$(PLATFORMS):
	$(info Building $@...)	
	GOOS=$(os) GOARCH=$(arch) go build -o bin/a9s-$(os)-$(arch) main.go

test:
	go test ./...

# Further reading: https://vic.demuzere.be/articles/golang-makefile-crosscompile/