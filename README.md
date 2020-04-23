Package **ipn** is an application-level library for working with a [Tailscale](https://tailscale.com) network. It can do the following:

1. Get your machines tailscale interface
2. Startup a listener directly on your machines tailscale interface
3. Efficiently query for the list of other machines on your tailscale network

Combining #2 and #3 together makes for a great authentication scheme for incoming web requests. This is great for admin panels, and other services that you only want to be accessed over the Tailscale network.

**Usage**:

Secure HTTP Server:

```
host, _, err := ipn.NetInterface()
if err != nil || host == nil {
  log.Fatalf("failed to get tailscale interface, maybe not on tailscale?")
}
s := &http.Server{
  Addr: net.JoinHostPort(host.String(), "80"),
  // refresh peer list every five minutes
  Handler: HTTPAuther(<other handler>, time.Minute * 5)
}
```

Note that if you are not listening explicitly on your tailscale interface, this will not be secure. This is because some european countries share the same ip ranges as tailscale uses, and could overlap.

Auto Peer Discovery:

```
me, err := ipn.Me()
if err != nil {
  log.Fatalf("failed to find myself...: %v", err)
}
q := ipn.DefaultQueryer()
peers, err := q.Query()
if err != nil {
  log.Fatalf("failed to query for peers: %v", err)
}
for _, p := range peers {
  log.Printf("found peer with hostname %s, addr %s and OS %s", p.Hostname, p.Addr, p.OS)
}
```

Please contribute if you feel like you want to!
