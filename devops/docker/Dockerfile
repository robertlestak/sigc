FROM golang:1.18 as builder

WORKDIR /src

COPY . .

RUN go build -o /bin/sigc cmd/sigc/*.go

FROM debian:bullseye as runtime

RUN apt-get update -y && apt-get install -y ca-certificates

COPY --from=builder /bin/sigc /bin/sigc

ENTRYPOINT [ "/bin/sigc" ]