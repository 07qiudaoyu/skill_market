# skillctl 构建脚本。
#
#   make build   只出当前平台的二进制（开发用，带版本号）
#   make dist    交叉编译 6 个平台到 dist/，并生成校验和（发布的是一整个目录）
#   make check   fmt + vet + build 三连，提交前跑一下
#
# 版本号注入：-X go_mod/oneself.Version=... 直接改 package 里的变量，
# 不用在源码里手动改版本，也不用生成别的文件。
BINARY  := skillctl
VERSION ?= 1.0.0
LDFLAGS := -s -w -X go_mod/oneself.Version=$(VERSION)
DIST    := dist

# 六个发布目标。CGO_ENABLED=0 保证静态链接，拿过去就能跑，不依赖目标机器的 libc。
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: build dist check fmt vet test clean dist-clean

# CGO_ENABLED=0 必须跟 dist 一致：默认（=1）会链接 cgo，同一个平台
# make build 和 make dist 出的二进制字节不同，本地核对的 sha 跟发布包对不上。
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

dist: dist-clean
	@mkdir -p $(DIST)
	@set -e; for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		out="$(DIST)/$(BINARY)-$(VERSION)-$${os}-$${arch}$${ext}"; \
		echo "  $$out"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 \
			go build -trimpath -ldflags "$(LDFLAGS)" -o "$$out" .; \
	done
	@cd $(DIST) && shasum -a 256 * > checksums.txt && echo "  $(DIST)/checksums.txt"

check: fmt vet build

fmt:
	gofmt -l . | grep . && echo "上面这些文件需要 gofmt -w" && exit 1 || true

vet:
	go vet ./...

test:
	go test ./...

clean:
	go clean
	rm -f $(BINARY)

dist-clean:
	rm -rf $(DIST)
