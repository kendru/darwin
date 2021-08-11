---
tags: []
created: Tue Aug 10 23:55:31 MDT 2021
---

# Linux Networking

## Enabling public traffic for local services.

When I was setting up a Rocky Linux server, I ran into a problem with a service that bound itself to `localhost:4000`. I wanted to connect to this server from another machine on the network, but incoming traffic came over the ethernet interface (a 192.168.* address), and the service was bound to the loopback. The solution was simply to enable the `net.ipv4.conf.all.route_localnet` setting:

```
sudo sysctl -w net.ipv4.conf.all.route_localnet=1
```

This allows traffic coming in on any ipv4 interface to be routed to loopback.

