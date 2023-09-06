# socks5
[中文文档](doc/README_zh.md)<br>
A lightweight SOCKS proxy server that supports socks4, socks4a, and socks5 protocols. The code is simple and easy to read, just like the original SOCKS protocol.

## Feature
* support [socks4](doc/SOCKS4.protocol.txt),[socks4a](doc/socks4A.protocol.txt),[socks5(TCP&UDP)](doc/rfc1928.txt)
* Supports [socks5 username/password authentication](doc/rfc1929.txt)

## Usage
Download the latest program for your operating system and architecture from the [Release](https://github.com/0990/socks5/releases) page.
Extracting,then execute the binary file directly (Linux platform requires execution permission)<br>
```bash
./ss5
```
or
```bash
./ss5 -c ./ss5.json
```
[Docker installation](doc/docker.md)

### Configuration
The ss5.json file in the extracted directory is the configuration file<br>
Simple configuration instructions:  
 ```
  ListenPort The listening port for TCP and UDP proxies, default is 1080
  UserName,Password Fill in if username/password authentication is required, default is empty
  LogLevel Log level (debug, info, warn, error)
``` 
[Advanced configuration](doc/config.md)
## Package Usage
```
go get github.com/0990/socks5  
```
Here is a simple example:
```
s := socks5.NewServer(socks5.ServerCfg{
	    ListenPort: 1080,
	    UserName:   "",
	    Password:   "",
	    UDPTimout:  60,
	    TCPTimeout: 60,
	    LogLevel:"error"
})
err := s.Run()
if err != nil {
	log.Fatalln(err)
}
```
## TODO
* Support BIND command

## Thanks
[txthinking/socks5](https://github.com/txthinking/socks5)  

