# Build the manager binary
FROM golang:1.22-alpine as builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /go/src/manager

RUN apk add -u -t build-tools curl git

# Create go cache
COPY go.mod go.sum ./
RUN --mount=type=ssh --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o manager cmd/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/src/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
