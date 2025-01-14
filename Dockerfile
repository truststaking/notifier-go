FROM golang:alpine as builder

WORKDIR /multiversx
COPY . .

WORKDIR /multiversx/cmd/notifier

RUN go build -o notifier

# ===== SECOND STAGE ======
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y openssl ca-certificates

COPY --from=builder /multiversx/cmd/notifier /multiversx

EXPOSE 5000

WORKDIR /multiversx

ENTRYPOINT ["./notifier"]
