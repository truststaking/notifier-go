FROM golang:1.18 as builder

WORKDIR /elrond
COPY . .

WORKDIR /elrond/cmd/notifier

RUN go build -o notifier

# ===== SECOND STAGE ======
FROM ubuntu:20.04
COPY --from=builder /elrond/cmd/notifier /elrond

EXPOSE 5000
EXPOSE 6380 
EXPOSE 5672
EXPOSE 15672

WORKDIR /elrond

ENTRYPOINT ["./notifier"]
CMD ["--api-type", "rabbit-api"]
