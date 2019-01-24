# Makefile to build argo-cd components
# Refer to this: https://github.com/argoproj/argo-cd/blob/v0.11.1/Makefile

PACKAGE=github.com/argoproj/argo-cd
CURRENT_DIR=$(shell pwd)
CLI_NAME=argocd
DIST_DIR=${CURRENT_DIR}/dist
VERSION=$(shell sed "s/\-[0-9]*//" ${CURRENT_DIR}/TAG )
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)
GIT_TREE_STATE=$(shell if [ -z "`git status --porcelain`" ]; then echo "clean" ; else echo "dirty"; fi)
PACKR_CMD=$(shell if [ "`which packr`" ]; then echo "packr"; else echo "go run vendor/github.com/gobuffalo/packr/packr/main.go"; fi)

override LDFLAGS += \
  -X ${PACKAGE}.version=${VERSION} \
  -X ${PACKAGE}.buildDate=${BUILD_DATE} \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \
  -X ${PACKAGE}.gitTreeState=${GIT_TREE_STATE}

.PHONY: cli
cli:
	${PACKR_CMD} build -v -i -ldflags '${LDFLAGS} -extldflags "-static"' -o ${DIST_DIR}/${CLI_NAME} ./cmd/argocd

# NOTE: we use packr to do the build instead of go, since we embed swagger files and policy.csv
# files into the go binary
.PHONY: server
server:
	${PACKR_CMD} build -v -i -ldflags '${LDFLAGS}' -o ${DIST_DIR}/argocd-server ./cmd/argocd-server

.PHONY:
repo-server:
	go build -v -i -ldflags '${LDFLAGS}' -o ${DIST_DIR}/argocd-repo-server ./cmd/argocd-repo-server

.PHONY:
controller:
	${PACKR_CMD} build -v -i -ldflags '${LDFLAGS}' -o ${DIST_DIR}/argocd-application-controller ./cmd/argocd-application-controller

.PHONY: argocd-util
argocd-util:
	# Build argocd-util as a statically linked binary, so it could run within the alpine-based dex container (argoproj/argo-cd#844)
	CGO_ENABLED=0 go build -v -i -ldflags '${LDFLAGS} -extldflags "-static"' -o ${DIST_DIR}/argocd-util ./cmd/argocd-util
