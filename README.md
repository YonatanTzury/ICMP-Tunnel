# ICMP Tunnel for Networking Class
The projects includes 2 parts:

1. Client - ...
1. Server - ...

### The server architecure is as follow

### Dialer
Each dialer store the following data:
1. icmpServer (string) - The ip to send to ICMP packets to
1. connectionCounter (int) - The amount of connection established since the client start
1. sleepDuration (int) - Duration to sleep between pooling
1. channelsBufferSize (int) - ????

#### General architecure
The ICMP transport use "pulling" architecure.

#### Connection Establish
The IMCP transport use 2 way handshake establish protocol.

### Listener
The listener is listenning on a single raw socket with the BPF filter
```go
"icmp[icmptype] == 8 and src host %s" // ICMPTypeEcho=8
```
Eeach packet receive is handle in the following way: