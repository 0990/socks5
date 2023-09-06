## Advanced Configuration
### UDPAdvertisedIP
The UDP proxy process for socks5 is as follows：
1. Establish a TCP proxy connection
2. Send the UDP proxy address through the TCP connection
3. The client then connects to this proxy address
```
    "ListenPort": 1080,
    "UDPAdvertisedIP": "192.168.0.100"
```
In the above configuration<br>
The UDP proxy listen addr is： 0.0.0.0:1080<br>
The advertised UDP proxy address is: 192.168.0.100:1080<br>
Please ensure that the UDP proxy can be connected using the advertised UDP proxy address.<br>
<mark>If UDPAdvertisedIP is empty, the local address will be used. In general, this is not a problem, but it needs to be configured when the local address cannot be obtained in environments like Docker.</mark>
### Custom UDP and TCP listening addresses
```
    "TCPListen": "127.0.0.1:1081",
    "UDPListen": "127.0.0.1:1082",
```
<mark>When TCPListen or UDPListen has a value, ListenPort will be ignored.</mark><br>
In the above scenario, the value of ListenPort, which is 1080, is invalid.<br>
<mark>Note: Improper configuration of UDPListen and UDPAdvertisedIP can cause UDP proxy failure.</mark>

### TCP and UDP Timeout
```
    "UDPTimout": 60,
    "TCPTimeout": 60,
```
In general, there is no need to change these values.