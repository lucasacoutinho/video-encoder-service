FROM golang:1.20-alpine3.18

# FFMPEG
RUN apk add --update ffmpeg bash curl make

# BASH
ENV PATH="$PATH:/bin/bash"

# Install Bento
ENV BENTO4_VERSION=1.6.0-637
ENV BENTO4_INSTALL_DIR=/opt/bento4
ENV BENTO4_BIN="/opt/bento4/bin"
ENV PATH="$PATH:/opt/bento4/bin"

RUN apk update \
  && apk add --no-cache \
  ca-certificates bash python3 wget libgcc make gcc g++

RUN cd /tmp/ \
  && wget -O Bento4-${BENTO4_VERSION}.tar.gz https://github.com/axiomatic-systems/Bento4/archive/v${BENTO4_VERSION}.tar.gz \
  && tar -xzvf Bento4-${BENTO4_VERSION}.tar.gz && rm Bento4-${BENTO4_VERSION}.tar.gz

RUN mkdir -p \
  ${BENTO4_INSTALL_DIR}/bin \
  ${BENTO4_INSTALL_DIR}/scripts \
  ${BENTO4_INSTALL_DIR}/include

RUN cd /tmp/Bento4-${BENTO4_VERSION}/Build/Targets/x86-unknown-linux \
  && make AP4_BUILD_CONFIG=Release

RUN cd /tmp \
  && cp -r Bento4-${BENTO4_VERSION}/Build/Targets/x86-unknown-linux/Release/. ${BENTO4_INSTALL_DIR}/bin \
  && cp -r Bento4-${BENTO4_VERSION}/Source/Python/utils/. ${BENTO4_INSTALL_DIR}/utils \
  && cp -r Bento4-${BENTO4_VERSION}/Source/Python/wrappers/. ${BENTO4_INSTALL_DIR}/bin \
  && cp -r Bento4-${BENTO4_VERSION}/Source/C++/**/*.h . ${BENTO4_INSTALL_DIR}/include

RUN rm -rf /var/cache/apk/* /tmp/*

WORKDIR /go/src

ENTRYPOINT ["tail", "-f", "/dev/null"]
