import test from "node:test";
import assert from "node:assert/strict";

import {
  countEndpointsInCIDRs,
  hostFromEndpoint,
  ipv4ToUint,
  isIPv4InCIDR,
  isPublicIPv4,
} from "./net.ts";

test("ipv4ToUint parses valid literals and rejects malformed ones", () => {
  assert.equal(ipv4ToUint("0.0.0.0"), 0);
  assert.equal(ipv4ToUint("255.255.255.255"), 4294967295);
  assert.equal(ipv4ToUint("192.168.1.1"), 3232235777);
  assert.equal(ipv4ToUint("256.1.1.1"), null, "octet > 255 must be rejected");
  assert.equal(ipv4ToUint("1.2.3"), null);
  assert.equal(ipv4ToUint("example.com"), null);
  assert.equal(ipv4ToUint(""), null);
});

test("isPublicIPv4 excludes private, loopback, link-local and reserved ranges", () => {
  for (const host of ["8.8.8.8", "1.1.1.1", "203.0.113.7", "172.32.0.1", "100.128.0.1", "223.255.255.255"]) {
    assert.equal(isPublicIPv4(host), true, `${host} should be public`);
  }
  for (const host of [
    "0.0.0.0",
    "10.0.0.1",
    "127.0.0.1",
    "169.254.1.1",
    "172.16.0.1",
    "172.31.255.255",
    "192.168.0.1",
    "100.64.0.1",
    "224.0.0.1",
    "999.1.1.1",
    "not-an-ip",
  ]) {
    assert.equal(isPublicIPv4(host), false, `${host} must not be public`);
  }
});

test("isIPv4InCIDR matches network membership", () => {
  assert.equal(isIPv4InCIDR("52.0.0.0", "52.0.0.0/11"), true);
  assert.equal(isIPv4InCIDR("52.31.255.255", "52.0.0.0/11"), true, "/11 covers 52.0.0.0-52.31.255.255");
  assert.equal(isIPv4InCIDR("52.32.0.0", "52.0.0.0/11"), false);
  assert.equal(isIPv4InCIDR("53.0.0.1", "52.0.0.0/11"), false);
  assert.equal(isIPv4InCIDR("1.2.3.4", "0.0.0.0/0"), true, "/0 covers everything");
  assert.equal(isIPv4InCIDR("1.2.3.4", "1.2.3.4/32"), true);
  assert.equal(isIPv4InCIDR("1.2.3.5", "1.2.3.4/32"), false);
  assert.equal(isIPv4InCIDR("1.2.3.4", "garbage"), false);
  assert.equal(isIPv4InCIDR("1.2.3.4", "52.0.0.0/33"), false);
  assert.equal(isIPv4InCIDR("nope", "52.0.0.0/11"), false);
});

test("hostFromEndpoint strips ports and unwraps IPv6 brackets", () => {
  assert.equal(hostFromEndpoint("198.51.100.20:443"), "198.51.100.20");
  assert.equal(hostFromEndpoint("198.51.100.20"), "198.51.100.20");
  assert.equal(hostFromEndpoint("[2001:db8::1]:443"), "2001:db8::1");
  assert.equal(hostFromEndpoint(""), "");
  assert.equal(hostFromEndpoint("   "), "");
});

test("countEndpointsInCIDRs counts unique public hosts only", () => {
  const hosts = [
    "52.0.0.5",
    "52.0.0.5", // duplicate endpoint
    "8.8.8.8", // public but outside the CIDR
    "127.0.0.1", // private: ignored
    "not-an-ip", // ignored
  ];
  const result = countEndpointsInCIDRs(hosts, ["52.0.0.0/11"]);
  assert.equal(result.total, 2, "only unique public hosts count");
  assert.equal(result.inside, 1);
});

test("countEndpointsInCIDRs reports zero coverage when nothing matches", () => {
  const result = countEndpointsInCIDRs(["8.8.8.8"], ["52.0.0.0/11"]);
  assert.deepEqual(result, { total: 1, inside: 0 });
});
