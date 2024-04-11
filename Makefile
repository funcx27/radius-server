.PHONY: ocserv radius
REPO=registry.kubeease.cn/tmp

ocserv:
	docker build -t ${REPO}/ocserv:1.1.6 docker-ocserv
radius:
	CGO_ENABLED=0 go build -o .output/radius-server
	docker build -t ${REPO}/radius-server .
test:
	docker rm -f ocserv radius
	docker run -d --name ocserv --privileged -e AUTH=RADIUS -e RADIUS_SERVER=172.30.179.201 -p:443:443 -e RADIUS_CLIENT_ID=test  -e AUTH=RADIUS ${REPO}/ocserv:1.1.6
	docker run -d --name radius --network=host ${REPO}/radius-server  -bypass 111111

push:
	docker push ${REPO}/ocserv:1.1.6
	docker push ${REPO}/radius-server