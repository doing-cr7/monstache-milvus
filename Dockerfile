FROM golang:1.25-alpine3.22 AS build
WORKDIR /src
COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOPROXY=https://goproxy.cn,direct

RUN go mod download && \
    go mod tidy -v && \
    go build -o /out/monstache .

FROM alpine:3.22
RUN apk --no-cache add ca-certificates && \
    mkdir -p /etc/monstache
COPY --from=build /out/monstache /bin/
COPY config.toml /etc/monstache/config.toml
ENTRYPOINT ["/bin/monstache", "-f", "/etc/monstache/config.toml"]