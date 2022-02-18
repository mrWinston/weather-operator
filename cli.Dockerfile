FROM quay.io/podman/stable:v3.4.4

ARG OPERATOR_SKD_VERSION=1.17.0
ENV OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v$OPERATOR_SKD_VERSION \
    OS=linux \
    ARCH=amd64



WORKDIR /workdir
RUN curl -LO $OPERATOR_SDK_DL_URL/operator-sdk_linux_amd64 && \
    gpg --keyserver keyserver.ubuntu.com --recv-keys 052996E2A20B5C7E && \
    curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt && \
    curl -LO ${OPERATOR_SDK_DL_URL}/checksums.txt.asc && \
    gpg -u "Operator SDK (release) <cncf-operator-sdk@cncf.io>" --verify checksums.txt.asc && \
    grep operator-sdk_${OS}_${ARCH} checksums.txt | sha256sum -c - && \
    chmod +x operator-sdk_${OS}_${ARCH} && \
    mv operator-sdk_${OS}_${ARCH} /bin/operator-sdk

WORKDIR /work

RUN rm -rf /workdir

USER podman

ENTRYPOINT operator-sdk
    

