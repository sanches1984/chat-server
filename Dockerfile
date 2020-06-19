FROM golang:alpine as builder
RUN apk update && \
    mkdir /app
ADD . /app/
WORKDIR /app
COPY ./ ./

RUN go install -v ./listener
RUN go build -o ./bin/chat ./listener/main.go

FROM alpine
RUN apk update && \
    adduser -D -H -h /app admin && \
    mkdir -p /app  && \
    chown -R admin:admin /app

USER admin
COPY --chown=admin --from=builder /app/bin/chat /app
COPY --chown=admin --from=builder /app/config.yml /app
WORKDIR /app

CMD ["/app/chat"]
