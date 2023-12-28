app := devspace
build_dir := build
dist_dir := dist

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
target_dir = '$(build_dir)/$(os)-$(arch)'
archive = $(dist_dir)/$(app)-$(os)-$(arch).tar.gz

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	md5 := md5
else
	md5 := md5sum
endif

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

build: core/wire_gen.go $(PLATFORMS)


$(PLATFORMS):
	@echo building $(os)/$(arch)...
	@mkdir -p $(target_dir)
	@mkdir -p $(dist_dir)
	@GOOS=$(os) GOARCH=$(arch) go build -ldflags "-X main.version=`cat version`" -o $(target_dir)/$(app)
	@tar zcf $(dist_dir)/$(app)-$(os)-$(arch).tar.gz -C $(target_dir) $(app)
	@cd $(dist_dir); $(md5) $(app)-$(os)-$(arch).tar.gz >> checksums.txt


core/wire_gen.go: core/wire.go
	@echo 'generate wire codes'
	@go run github.com/google/wire/cmd/wire ./...

debug-build: core/wire_gen.go
	@go build -gcflags="all=-N -l" -ldflags "-X main.version=`cat version`" -o $(app)

lint: core/wire_gen.go
	@echo running golangcli-lint...
	@golangci-lint run
	@echo

test: core/wire_gen.go
	@echo testing...
	@go test -timeout 10s -v ./...
	@echo

clean:
	@go clean
	@go clean -testcache
	@rm -rf build
	@rm -rf dist
	@rm -f core/wire_gen.go
	@rm -f $(app)

prerelease: vet test

tag:
	git tag `cat version`
	git push origin `cat version`

all: build

.PHONY: build clean test vet wire generate prerelease $(PLATFORMS)
