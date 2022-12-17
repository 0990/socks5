## 高级配置
### UDP下发地址（UDPAdvertisedIP）
socks5的UDP代理原理是：
1. 先建立tcp代理连接
2. 通过tcp连接下发UDP的代理地址<br>
3. 客户端再连接这个代理地址
```
    "ListenPort": 1080,
    "UDPAdvertisedIP": "192.168.0.100"
```
上面配置的情况下<br>
UDP代理地址为： 0.0.0.0:1080<br>
下发的UDP代理地址为： 192.168.0.100:1080<br>
请确保这个通过下发的UDP代理地址能连上这个UDP代理<br>
<mark>UDPAdvertisedIP为空时，则使用本地地址，一般情况下没有问题，但在docker等环境获取不到本地地址时，需要配置此项</mark>
### 自定义UDP,TCP监听地址
```
    "TCPListen": "127.0.0.1:1081",
    "UDPListen": "127.0.0.1:1082",
```
<mark>当TCPListen或UDPListen有值时，ListenPort则会被忽略</mark><br>
上面的情况下，ListenPort的值1080无效

<mark>注意：UDPListen、UDPAdvertisedIP配置不当，会导致UDP代理不了</mark>

### TCP,UDP超时时间
```
    "UDPTimout": 60,
    "TCPTimeout": 60,
```
这个一般情况下不用更改