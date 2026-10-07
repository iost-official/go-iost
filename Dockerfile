FROM ubuntu:22.04

# Install project
WORKDIR /workdir
COPY target/iserver target/iwallet target/itest ./
COPY vm/v8vm/v8/libjs/ ./vm/v8vm/v8/libjs/
COPY vm/v8vm/v8/libv8/_linux_amd64/ /lib/x86_64-linux-gnu/
COPY config/docker/iserver.yml /var/lib/iserver/
COPY config/genesis/ /var/lib/iserver/genesis/

ENV PATH="/workdir:${PATH}"
CMD ["iserver", "-f", "/var/lib/iserver/iserver.yml", "2>&1"]
