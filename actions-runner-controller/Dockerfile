FROM quay.io/cybozu/golang:1.15-focal AS build

WORKDIR /workspace

ARG ACTIONS_RUNNER_CONTROLLER_VERSION=0.17.0

RUN git clone -b v${ACTIONS_RUNNER_CONTROLLER_VERSION} https://github.com/summerwind/actions-runner-controller

RUN cd actions-runner-controller && \
  go mod download && \
  CGO_ENABLED=0 go build -a -o manager main.go

FROM quay.io/cybozu/ubuntu:20.04

WORKDIR /

COPY --from=build /workspace/actions-runner-controller/manager .

USER 10000:10000

ENTRYPOINT ["/manager"]
