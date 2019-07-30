ARG BUILDER_IMG=golang:alpine

FROM ${BUILDER_IMG} AS builder
LABEL stage=builder
ADD . /src
RUN apk add git
RUN cd /src && go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /src/quiteabot /usr/local/bin/quiteabot
CMD /usr/local/bin/quiteabot
