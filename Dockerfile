FROM golang:1.19 AS base

RUN apt-get update \
  && apt-get -y install make openssh-client ca-certificates && update-ca-certificates

# Install unzip
RUN apt-get update && apt-get install -y unzip

FROM base AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY Makefile .
COPY .git ./.git
COPY . ./
RUN make build

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/template-service /template-service

ENTRYPOINT ["/templatey-service", "serve"]
