FROM golang:1.22 as builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
WORKDIR /todo

COPY go.mod go.sum ./
RUN go mod download -x

COPY main.go ./
RUN go build -o app main.go

FROM scratch
COPY --from=builder /todo/app /
EXPOSE 443
ENTRYPOINT [ "/app" ]
