FROM ubuntu:22.04

LABEL maintainer="modelz-support@tensorchord.ai"
RUN apt-get -qq update \
    && apt-get -qq install -y --no-install-recommends ca-certificates curl

COPY agent /usr/bin/agent
ENTRYPOINT ["/usr/bin/agent"]
