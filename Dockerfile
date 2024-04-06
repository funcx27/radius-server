FROM alpine

COPY .output/radius-server  /usr/local/bin/radius-server
WORKDIR /radius-server
COPY userconfig.yaml  .
ENTRYPOINT [ "radius-server" ]