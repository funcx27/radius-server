#!/bin/bash
set -e

version=$(ocserv -v 2>&1  | grep ocserv | awk '{print $NF}')
certs_create(){
if [ ! -f /etc/ocserv/certs/server-key.pem ] || [ ! -f /etc/ocserv/certs/server-cert.pem ]; then
		# Check environment variables
		if [ -z "$CA_CN" ]; then
			CA_CN="VPN CA"
		fi

		if [ -z "$CA_ORG" ]; then
			CA_ORG="Big Corp"
		fi

		if [ -z "$CA_DAYS" ]; then
			CA_DAYS=9999
		fi

		if [ -z "$SRV_CN" ]; then
			SRV_CN="www.example.com"
		fi

		if [ -z "$SRV_ORG" ]; then
			SRV_ORG="MyCompany"
		fi

		if [ -z "$SRV_DAYS" ]; then
			SRV_DAYS=9999
		fi

# No certification found, generate one
mkdir -p /etc/ocserv/certs
cd /etc/ocserv/certs
certtool --generate-privkey --outfile ca-key.pem
cat > ca.tmpl <<-EOCA
cn = "$CA_CN"
organization = "$CA_ORG"
serial = 1
expiration_days = $CA_DAYS
ca
signing_key
cert_signing_key
crl_signing_key
EOCA
certtool --generate-self-signed --load-privkey ca-key.pem --template ca.tmpl --outfile ca.pem
certtool --generate-privkey --outfile server-key.pem 
cat > server.tmpl <<-EOSRV
cn = "$SRV_CN"
organization = "$SRV_ORG"
expiration_days = $SRV_DAYS
signing_key
encryption_key
tls_www_server
EOSRV
certtool --generate-certificate --load-privkey server-key.pem --load-ca-certificate ca.pem --load-ca-privkey ca-key.pem --template server.tmpl --outfile server-cert.pem
fi
}

network_config(){
	# Open ipv4 ip forward
	sysctl -w net.ipv4.ip_forward=1

	# Enable NAT forwarding
	iptables -t nat -A POSTROUTING -j MASQUERADE
	iptables -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

	# Enable TUN device
	mkdir -p /dev/net
	if [ ! -c /dev/net/tun ];then
	mknod /dev/net/tun c 10 200
	chmod 600 /dev/net/tun
	fi
}
config_file_init(){
	echo "$RADIUS_SERVER $RADIUS_SECRET" >> /etc/radcli/servers
	sed -i "s/^authserver.*/authserver $RADIUS_SERVER/; s/^acctserver.*/acctserver $RADIUS_SERVER/" /etc/radcli/radiusclient.conf
	sed '/^#/d; /^$/d' /opt/ocserv-"$version".conf > /opt/ocserv.conf
	echo -e "\n#####addon configs####\ncompression = true\nserver-cert = /etc/ocserv/certs/server-cert.pem\nserver-key = /etc/ocserv/certs/server-key.pem" >> /opt/ocserv.conf
	if [ "$DTLS" = disabled ];then
		sed -i '/^udp-port.*/d' /opt/ocserv.conf
	fi
	if [ ! -f /etc/ocserv/ocserv.conf ];then
		if [ "$AUTH" = "RADIUS" ];then
		    if [ -n "$RADIUS_CLIENT_ID" ];then
			  nas_id=,nas-identifier=$RADIUS_CLIENT_ID
			fi
			sed -e "s|auth =.*|auth = \"radius[config=/etc/radcli/radiusclient.conf,groupconfig=true${nas_id}]\"|" \
			/opt/ocserv.conf  >  /etc/ocserv/ocserv.conf
			if [ "$RADIUS_CLIENT_ACCT" = enabled  ];then
				sed -i '1a\acct = "radius[config=/etc/radcli/radiusclient.conf]"' /etc/ocserv/ocserv.conf
			fi
		else 
			echo -e "config-per-group = /etc/ocserv/config-per-group\ndefault-group-config = /etc/ocserv/config-per-group/default.conf\nconfig-per-user = /etc/ocserv/config-per-user\ndefault-user-config = /etc/ocserv/config-per-group/default.conf" >> /opt/ocserv.conf
			sed 's|auth =.*|auth = "plain[passwd=/etc/ocserv/ocpasswd,otp=/etc/ocserv/users.otp]"|' /opt/ocserv.conf  >  /etc/ocserv/ocserv.conf
		    mkdir -p /etc/ocserv/config-per-group /etc/ocserv/config-per-user
			if  [ ! -f /etc/ocserv/config-per-group/default.conf ];then
				echo route = 192.168.99.1/255.255.255.255 > /etc/ocserv/config-per-group/default.conf
			fi
			if  [ ! -f /etc/ocserv/config-per-user/default.conf ];then
				echo route = 192.168.99.1/255.255.255.255 > /etc/ocserv/config-per-user/default.conf
			fi
			touch /etc/ocserv/ocpasswd
				# Create a test user
				#if [ -z "$NO_TEST_USER" ] && [ ! -f /etc/ocserv/ocpasswd ]; then
					#echo "Create test user 'test' with password 'test'"
					#echo 'test:Route,All:$5$DktJBFKobxCFd7wN$sn.bVw8ytyAaNamO.CvgBvkzDiFR6DaHdUzcif52KK7' > /etc/ocserv/ocpasswd
				#fi
		fi
	fi

}
certs_create
config_file_init
network_config

echo -e "\n############################## /etc/ocserv/ocserv.conf########################################"
cat /etc/ocserv/ocserv.conf
echo -e "############################## /etc/ocserv/ocserv.conf########################################\n"
# Run OpennConnect Server
exec "$@"
