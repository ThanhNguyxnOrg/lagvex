//go:build !windows

package client

import (
	"errors"
	"net/netip"
)

var errRoutesNotSupported = errors.New("Windows RouteManager is only supported on Windows")

type RouteManager struct{}

func NewRouteManager() *RouteManager {
	return &RouteManager{}
}

func (r *RouteManager) ConfigureAdapter(tunIfIndex uint32, innerIP netip.Addr, prefixLen, mtu int) error {
	return errRoutesNotSupported
}

func (r *RouteManager) PinRelayRoute(relayIP netip.Addr) error {
	return errRoutesNotSupported
}

func (r *RouteManager) InstallGameRoutes(cidrs []string) {}

func (r *RouteManager) RemoveGameRoutes() {}

func (r *RouteManager) RemoveAll() {}

func (r *RouteManager) ActiveRouteCount() int { return 0 }
