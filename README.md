# socks5
轻量的socks代理服务器，支持socks4,socks4a,socks5,代码简单易读，就像sock原始协议一样

## Feature
* 支持[socks4](doc/SOCKS4.protocol.txt),[socks4a](doc/SOCKS4A.protocol.txt),[socks5(TCP&UDP)](doc/rfc1928.txt)
* 支持[socks5用户名密码鉴权](doc/rfc1929.txt)

## 使用
 [下载页面](https://github.com/0990/socks5/releases) </br>
 根据平台选择下载文件，解压后直接执行二进制文件即可（linux平台需要加执行权限)

### 配置
 解压后同目录下的ss5.json是配置文件，各项配置字段说明如下  
 ```
  ListenPort 监听端口，默认是1080  
  UserName,Password 需要用户名密码鉴权时填写，默认为空
  UDPTimeout udp读超时时间，默认60s
  TCPTimeout tcp读超时时间，默认60s
  UDPAddr udp代理的地址
  LogLevel 日志等级（debug,info,warn,error)
``` 

## 示例
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
	    LogLevel:"error"
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

