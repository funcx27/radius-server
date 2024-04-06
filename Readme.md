## radius

- RADIUS Authentication doc
  https://datatracker.ietf.org/doc/html/rfc2865

- RADIUS Accounting doc
  https://datatracker.ietf.org/doc/html/rfc2866

- ocserv 1.1.6 doc
  
  https://gitlab.com/openconnect/ocserv/-/blob/1.1.6/doc/README-radius.md
  https://gitlab.com/openconnect/ocserv/-/blob/1.1.6/doc/sample.config?ref_type=tags


## test

```sh
cat > userconfig.yaml <<EOF
#通过radius下发路由配置
routes:
  test1:
    - 192.168.111.0/24
    - 192.168.112.0/24
  test2:
    - 192.168.211.0/24
    - 192.168.212.0/24

users:
  69789@test: [test1,test2]
  22222@default: [test2]
  33333@default: [test3]


# 通过radius获取登录用户的ocserv group名称, 匹配ocserv本地的路由配置
#groups:
#  bangdao:
#    - 69789@test
#  test1:
#    - 11111@default
EOF


docker rm -f ocserv radius
docker run -d --name ocserv --privileged -e AUTH=RADIUS -e RADIUS_SERVER=localhost -p:443:443 -e RADIUS_CLIENT_ID=test  -e AUTH=RADIUS registry.kubeease.cn/tmp/ocserv:1.1.6
docker run -d --name radius -v$PWD/userconfig.yaml:/radius-server/userconfig.yaml --network=container:ocserv registry.kubeease.cn/tmp/radius-server -listenaddr 127.0.0.1:1812
docker logs -f radius
```