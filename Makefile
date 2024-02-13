PLATFORMS := windows/386 windows/amd64 darwin/amd64 darwin/arm64 linux/arm64 linux/amd64 linux/386

# Variables will be evaluated on use > this will also work within a loop
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

# Metadata to be carried into the binary enabling the user/dev to relate the binary to the documentation and source code
timestamp = $(shell date +%s)
version = "v0.10.0"
lastCommit = $(shell git log -1 --pretty=format:"%H")

# Your platform
build:
	$(info Build time is $(timestamp))
	$(info Version is $(version))
	$(info Last commit was $(lastCommit))
	go build -v -ldflags "-X 'github.com/anynines/a9s-cli-v2/cmd.BuildTimestamp="$(timestamp)"' -X 'github.com/anynines/a9s-cli-v2/cmd.CliVersion=$(version)' -X 'github.com/anynines/a9s-cli-v2/cmd.LastCommit=$(lastCommit)'" -o bin/a9s main.go
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