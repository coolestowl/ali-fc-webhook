GO     := go
ENTRY  := main.go
DIST   := .
OUTPUT := ${DIST}/main

GO_VERSION   := $(shell go version | awk '{print $$3}')
GIT_DESCRIBE := $(shell git describe --always --tags --dirty)
GIT_HASH     := $(shell git rev-parse HEAD)
CURRENT_TIME := $(shell date +'%Y-%m-%d %H:%M:%S')

GLOBAL_LD_FLAGS := -X 'github.com/coolestowl/ali-fc-webhook/build.Version=${GIT_DESCRIBE}' \
	-X 'github.com/coolestowl/ali-fc-webhook/build.GoVersion=${GO_VERSION}' \
	-X 'github.com/coolestowl/ali-fc-webhook/build.GitHash=${GIT_HASH}' \
	-X 'github.com/coolestowl/ali-fc-webhook/build.BuildTime=${CURRENT_TIME}'

RELEASE_OS   := linux #darwin windows
RELEASE_ARCH := amd64 #arm64

.PHONY: release
release:
	@for TARGET_OS in ${RELEASE_OS} ; do \
		for TARGET_ARCH in ${RELEASE_ARCH} ; do \
			echo "build for" $${TARGET_OS}/$${TARGET_ARCH} ; \
			if [ $${TARGET_OS} = 'windows' ]; then \
				FILE_EXT=.exe ; \
			fi ; \
			FILENAME=${DIST}/bin-$${TARGET_OS}-$${TARGET_ARCH}$${FILE_EXT} ; \
			EXTRA_FLAGS="-X 'github.com/coolestowl/ali-fc-webhook/build.OSArch=$${TARGET_OS}/$${TARGET_ARCH}'" ; \
			GOOS=$${TARGET_OS} GOARCH=$${TARGET_ARCH} ${GO} build -ldflags "${GLOBAL_LD_FLAGS} $${EXTRA_FLAGS}" -o $${FILENAME} ${ENTRY} ; \
		done \
	done
	@cd ${DIST} && for file in $$(ls); do shasum -a256 $${file} > $${file}.sha256sum; done && cd - > /dev/null
	@cd ${DIST} && for file in *.sha256sum; do shasum -c $${file}; done && cd - > /dev/null

.PHONY: build
build:
	@for _ in _ ; do \
		EXTRA_FLAGS="-X 'github.com/coolestowl/ali-fc-webhook/build.OSArch=linux/amd64'" ; \
		${GO} build -ldflags "${GLOBAL_LD_FLAGS} $${EXTRA_FLAGS}" -o ${OUTPUT} ${ENTRY} ; \
	done

.PHONY: build-static
build-static:
	@for _ in _ ; do \
		EXTRA_FLAGS="-X 'github.com/coolestowl/ali-fc-webhook/build.OSArch=linux/amd64'" ; \
		CGO_ENABLED=0 ${GO} build -ldflags "${GLOBAL_LD_FLAGS} $${EXTRA_FLAGS}" -o ${OUTPUT} ${ENTRY} ; \
	done

.PHONY: test
test:
	@${GO} test -v ./...

.PHONY: clean
clean:
	@-rm ${OUTPUT}
	@-rm ${DIST}/*
