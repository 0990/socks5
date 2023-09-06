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

## environment
```
PROXY_USER username
PROXY_PASSWORD password
PROXY_UDP_TIMEOUT tcp timeout seconds(default 60s)
PROXY_TCP_TIMEOUT udp timeout seconds(default 60s)
PROXY_PORT tcp and udp listen port(default 1080)
PROXY_ADVERTISED_UDP_IP udp advertised ip
```



