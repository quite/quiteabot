ARG BUILDER_IMG=golang:alpine

FROM ${BUILDER_IMG} AS builder
LABEL stage=builder
RUN apk add --no-cache git
WORKDIR /src
COPY . .
RUN go build

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /src/quiteabot /usr/local/bin/quiteabot
CMD /usr/local/bin/quiteabot
