## socks5
轻量的socks5代理服务器，代码简单易读，就像sock5原始协议一样([RFC1928](https://tools.ietf.org/html/rfc1928),
[RFC1929](https://tools.ietf.org/html/rfc1929))

## Feature
* 支持TCP/UDP
* 支持用户名密码鉴权

## Usage
 根据平台在此选择下载文件，解压后直接执行二进制文件即可（linux平台需要加执行权限)

### Config
 解压后同目录下的ss5.json是配置文件，各项配置字段说明如下  
 ```
  ListenPort 监听端口，默认是1080  
  UserName,Password 需要用户名密码鉴权时填写，默认为空
  UDPTimeout udp读超时时间，默认60s
  TCPTimeout tcp读超时时间，默认60s
  UDPAddr udp代理的地址
```
## Docker
docker run -d -p 1080:1080 0990/socks5

更多:
* 用户名密码鉴权  
docker run -d -p 1080:1080 -e PROXY_USER=XXX -e PROXY_PASSWORD=XXX 0990/socks5

* 支持udp  
docker run -d -p PORT_A:1080 -p PORT_B:1080/udp -e PROXY_UDP_ADDR=SERVER_IP:PORT_B 0990/socks5  

## Package Usage
```
go get github.com/0990/socks5  
```
以下为一个简单示用
```
s := socks5.NewServer(socks5.ServerCfg{
	    ListenPort: 1080,
	    UserName:   "",
	    Password:   "",
	    UDPTimout:  60,
	    TCPTimeout: 60,
	    UDPAddr:"127.0.0.1:1080",
})
err := s.Run()
if err != nil {
	log.Fatalln(err)
}
```
## TODO
* 支持 BIND 命令

## Thanks
[txthinking/socks5](https://github.com/txthinking/socks5)  

