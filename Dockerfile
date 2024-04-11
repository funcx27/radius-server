FROM debian:12-slim
RUN  sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/* &&\
     apt-get update &&\
     apt install -y ca-certificates procps &&\
     update-ca-certificates
COPY .output/radius-server .output/radius-client  /usr/local/bin/
WORKDIR /radius-server
COPY userconfig.yaml  .
ENV TZ=Asia/Shanghai
ENTRYPOINT [ "radius-server" ]