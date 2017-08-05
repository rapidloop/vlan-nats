
# vlan-nats

_Disclaimer: This is only a fun experiment. Do not use it for anything serious._

vlan-nats creates a virtual LAN using [NATS](https://nats.io/). Backed by a NATS
server (or cluster), vlan-nats can create and run a network interface that is
connected to a virtual L2 switch.

vlan-nats is written in entirely in Go. Currently, it works only on Linux.
Multicast is not supported.

First, get and build vlan-nats:

```
$ go get github.com/rapidloop/vlan-nats
```

Then, on each machine, do:

```
sudo vlan-nats -n nats://{MY-NATS-SERVER}:4222 &
sudo ip addr add 10.1.0.{CHANGEME} broadcast 10.1.255.255 dev vnats0
sudo ip link set vnats0 up
sudo ip route add 10.1.0.0/16 dev vnats0
```

You should replace `{CHANGEME}` with a unique number. The `{MY-NATS-SERVER}` is
the IP or host of a reachable NATS server (all machines must connect to the same
NATS cluster).

Congratulations! You now have all your machines reachable on the virtual subnet
10.1.0.0/16. You should be able to ping each other:

```
$ ip a show vnats0
7: vnats0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN group default qlen 1000
    link/ether 7e:8f:d4:17:2b:0c brd ff:ff:ff:ff:ff:ff
    inet 10.1.0.1/32 brd 10.1.255.255 scope global vnats0
       valid_lft forever preferred_lft forever
    inet6 fe80::7c8f:d4ff:fe17:2b0c/64 scope link tentative dadfailed
       valid_lft forever preferred_lft forever
$ ping 10.1.0.2
PING 10.1.0.2 (10.1.0.2) 56(84) bytes of data.
64 bytes from 10.1.0.2: icmp_seq=1 ttl=64 time=1.70 ms
64 bytes from 10.1.0.2: icmp_seq=2 ttl=64 time=1.60 ms
^C
--- 10.1.0.2 ping statistics ---
2 packets transmitted, 2 received, 0% packet loss, time 1003ms
rtt min/avg/max/mdev = 1.603/1.654/1.706/0.065 ms
$
```

Cool!

To cleanup, just `sudo pkill vlan-nats`. The interface is not persisted.

### How It Works

vlan-nats creates a TAP interface. All broadcast frames from the interface are
published to the NATS topic `vlan.{ID}` and unicast frames are published to
`vlan.{ID}.{DST_ETHADDR}`. The process subscribes to `vlan.{ID}` and
`vlan.{ID}.{OWN_ETHADDR}` and writes out any received frames into the TAP
interface.

Windows and OS X do not natively support TAP devices, but (free) 3rd party
drivers are available.

### Things To Try Out

* Windows and OS X can be supported with 3rd party TAP drivers.
* Run a DHCP server instead of assigning static IPs.
* Use a TLS-enabled, authenticated, public NATS server -- like a VPN!


