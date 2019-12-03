FROM alpine:latest
LABEL maintainer="roman.van-gemmeren@chefkoch.de"

RUN apk --no-cache --update add \
    ca-certificates

COPY ./gitlab-ci-pipelines-exporter_linux_amd64 .
ENTRYPOINT ["./gitlab-ci-pipelines-exporter_linux_amd64"]

