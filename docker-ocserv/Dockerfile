FROM debian:12-slim



RUN  sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/* &&\
     apt-get update
RUN apt-get install -y --no-install-recommends gnutls-bin procps ocserv iptables  oathtool net-tools iproute2 curl iputils-ping tzdata \
    && apt-get clean
     #yum install -y ocserv net-tools sssd-client pam_oath oathtool gen-oath-safe&&\
RUN set -x &&\
    sed -i "s@^ipv4-network.*@ipv4-network = 192.168.99.0@" /etc/ocserv/ocserv.conf &&\
    sed -i "s@^max-clients.*@max-clients = 100@" /etc/ocserv/ocserv.conf &&\
    sed -i "s@^log-level.*@log-level = 2@" /etc/ocserv/ocserv.conf &&\
    sed -i "/^route/d; /^dns/d" /etc/ocserv/ocserv.conf &&\
    version=$(ocserv -v 2>&1  | grep ocserv | awk '{print $NF}') &&\
    mv /etc/ocserv/ocserv.conf /opt/ocserv-$version.conf

COPY ocserv.pam  /etc/pam.d/ocserv
ENV AUTH=RADIUS \
    DTLS=disabled \
    RADIUS_CLIENT_ID=default \
    RADIUS_CLIENT_ACCT=disabled \
    RADIUS_SERVER=localhost\
    RADIUS_SECRET=test123
COPY docker-entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
WORKDIR /etc/ocserv
CMD ["ocserv", "-c", "/etc/ocserv/ocserv.conf", "-f"]
EXPOSE 443/tcp 443/udp

