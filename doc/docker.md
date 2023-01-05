# Docker image

## Run
```
docker run -d --name socks5 -p 1080:1080 0990/socks5:latest
```

## Run with udp support
```
docker run -d --name socks5 -p 1080:1080/tcp -p 1080:1080/udp -e PROXY_UDP_IP="x.x.x.x" 0990/socks5:latest
```
x.x.x.x为docker的访问ip

## 环境变量
```
PROXY_USER 用户名（鉴权用)
PROXY_PASSWORD 密码（鉴权用)
PROXY_UDP_TIMEOUT tcp超时时间（默认60s)
PROXY_TCP_TIMEOUT udp超时时间（默认60s)
PROXY_PORT tcp及udp代理端口(默认1080)
PROXY_UDP_IP udp下发地址
```



