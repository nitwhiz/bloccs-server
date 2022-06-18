FROM golang:1.18.3 as builder

WORKDIR /app

COPY ./ /app

RUN CGO_ENABLED=0 go build -o ./build/server ./cmd/server

FROM alpine:3.15.4

LABEL org.opencontainers.image.source="https://github.com/nitwhiz/bloccs-server"

COPY --from=builder /app/build/server /server

EXPOSE 7000

CMD [ "/server" ]
