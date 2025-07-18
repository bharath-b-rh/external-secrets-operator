# Build the external-secrets-operator binary
FROM docker.io/golang:1.23 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY . .

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o external-secrets-operator cmd/external-secrets-operator/main.go

# Use distroless as minimal base image to package the external-secrets-operator binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi9-minimal:9.4
WORKDIR /
COPY --from=builder /workspace/external-secrets-operator /bin/external-secrets-operator
USER 65534:65534

ENTRYPOINT ["/bin/external-secrets-operator"]
