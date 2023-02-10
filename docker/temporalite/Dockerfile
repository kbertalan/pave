FROM debian:11-slim AS builder

ARG VERSION=0.3.0

WORKDIR /opt/temporalite

RUN apt update \
    && apt install -y curl tar gzip
RUN curl -L https://github.com/temporalio/temporalite/releases/download/v${VERSION}/temporalite_${VERSION}_linux_amd64.tar.gz -o temporalite_${VERSION}_linux_amd64.tar.gz
RUN tar -xvzf temporalite_${VERSION}_linux_amd64.tar.gz

FROM gcr.io/distroless/base-debian11

COPY --from=builder /opt/temporalite/temporalite /
EXPOSE 7233
EXPOSE 8233

VOLUME /data

ENTRYPOINT ["/temporalite", "start", "-n", "default", "--ip" , "0.0.0.0", "-f", "/data/temporalite.db"]
