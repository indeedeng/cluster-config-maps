# Build the binary
FROM --platform=$BUILDPLATFORM golang:1.17 as builder

ARG BUILDPLATFORM
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY apis/ apis/
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -o ccm-csi-plugin cmd/ccm-csi-plugin/main.go

FROM alpine:3.14

RUN apk add --no-cache ca-certificates e2fsprogs findmnt

WORKDIR /
COPY --from=builder /workspace/ccm-csi-plugin .


ENTRYPOINT ["/ccm-csi-plugin"]
