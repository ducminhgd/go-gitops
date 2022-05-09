FROM golang:1.17-alpine as builder

ENV GOARCH="amd64" \
    GOOS=linux

WORKDIR /app
COPY . .
RUN go build -tags musl --ldflags "-extldflags -static"  -o gitlab ./cmd/gitlab && go build -tags musl --ldflags "-extldflags -static"  -o workloads ./cmd/workloads
RUN apk add --no-cache --update curl && curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/local/bin

FROM alpine:3.13.6

WORKDIR /

RUN apk add --no-cache --update curl

COPY --from=builder /app/gitlab /usr/local/bin/

# musl is a libc and it look for hostname resolving in `/etc/nsswitch.conf`. glibc is a torlerance libc, it has fail-over
#   options for hostname resolving. Since alpine is a musl-based linux and `/etc/nsswitch.conf` is removed from alpine,
#   so that `--add-host` arguments does not effect. That's why we need to create image's own nsswitch.conf file.
RUN echo 'hosts: files dns' > /etc/nsswitch.conf
