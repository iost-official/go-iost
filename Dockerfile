FROM ubuntu:22.04

# Install project
WORKDIR /workdir
COPY target/iserver target/iwallet target/itest ./

COPY config/docker/iserver.yml /var/lib/iserver/
COPY config/genesis/ /var/lib/iserver/genesis/

ENV PATH="/workdir:${PATH}"
CMD ["iserver", "-f", "/var/lib/iserver/iserver.yml", "2>&1"]
