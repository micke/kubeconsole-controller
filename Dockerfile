# Build
FROM golang:1.14.6-buster as builder

ENV USER=app
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /app/

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/kubeconsole-controller

# Run
FROM debian:buster AS runtime

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/kubeconsole-controller /kubeconsole-controller

USER app:app

ENTRYPOINT ["/kubeconsole-controller"]
