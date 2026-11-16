package client

// fullTunnelPrefixes are the two half-default routes used to send all IPv4
// traffic through the tunnel without deleting the real default route. When
// they are removed, the original default route takes over again immediately.
var fullTunnelPrefixes = []string{"0.0.0.0/1", "128.0.0.0/1"}
