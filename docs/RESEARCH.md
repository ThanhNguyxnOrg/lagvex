# LAGVEX — Networking Research & Architecture Handoff

> 🔗 [Back to Project README.md](../README.md) | [🤝 Contributor Guide](../CONTRIBUTING.md) | [Roadmap](ROADMAP.md) | [Architecture](ARCHITECTURE.md)
>
> **Purpose:** research-backed technical brief for an implementation agent building LAGVEX, an open-source Windows game network accelerator.
>
> **Verified research date:** 2026-09-12
>
> **Evidence rule:** Every factual technical claim in this document is tied to a primary source, official repository/documentation, RFC, or academic publication. Ideas marked **DERIVED DESIGN** are engineering proposals for LAGVEX and are **not** claims that a paper or project has already proven them for LAGVEX/FPS games.

---

## 0. Project definition

### Current LAGVEX concept

LAGVEX is intended to improve the network path of selected game traffic on Windows rather than optimize the entire machine/network blindly.

Current target architecture:

```text
Game process
    |
    | game traffic
    v
Windows routing / Wintun
    |
    | selected destination traffic
    v
LAGVEX tunnel client
    |
    | encrypted UDP transport
    v
Self-hosted VPS relay
    |
    | normal Internet routing
    v
Game server
```

### Core product goal

**Primary goal:** give game traffic a potentially better/stabler Internet path when the user's default ISP path is suboptimal.

**Do not claim:** that a relay always lowers latency. A relay can only help when the complete path through the relay is better than the direct path for the relevant traffic.

This follows from the basic path structure of a relay system and from research showing that Internet inter-domain routing can be circuitous and that latency-aware path selection is a legitimate optimization target. See [Latency-Aware Inter-domain Routing](https://arxiv.org/abs/2410.13019).

---

# 1. Evidence hierarchy

Use sources in this order when making implementation decisions:

1. **IETF RFC / standards:** protocol semantics and interoperability.
2. **Official project repositories/documentation:** concrete implementation behavior.
3. **Peer-reviewed conference/journal papers:** measured behavior and research results.
4. **Project README/blogs:** useful implementation context, but not equivalent to controlled academic evidence.
5. **Secondary articles / search summaries:** discovery only; do not use them as the basis for a core architectural claim when a primary source exists.

When a source describes cloud gaming rather than ordinary client-rendered multiplayer FPS, treat it as adjacent evidence, not direct proof for CS/Valorant/Apex-style traffic.

---

# 2. Existing architecture: what is technically well-founded

## 2.1 Wintun/TUN is a valid foundation

The official Wintun repository describes Wintun as a **Layer 3 TUN driver for Windows** intended for projects that need Layer 3 tunneling devices with primarily userspace implementations. It was created for WireGuard and is explicitly intended for wider use. [S1]

**Source:** WireGuard/Wintun official repository:
- https://github.com/WireGuard/wintun
- https://www.wintun.net/

### Implication for LAGVEX

Using Wintun as the virtual interface for LAGVEX is consistent with an established Windows userspace tunneling model.

**Evidence status:** DIRECTLY SUPPORTED by official Wintun documentation.

---

## 2.2 Windows route-table steering through a TUN interface is established

Xray's TUN implementation documents Windows support through Wintun and shows that traffic can be routed through the created adapter using Windows routes. The documentation explicitly warns that incorrectly routing traffic can create routing loops and network failure. [S2]

The current Xray Windows TUN implementation uses Wintun and `winipcfg`, and manages routes through the Windows networking stack. [S3]

**Sources:**
- https://github.com/XTLS/Xray-core/blob/main/proxy/tun/README.md
- https://github.com/XTLS/Xray-core/blob/main/proxy/tun/tun_windows.go

### Implication for LAGVEX

A selective routing design such as:

```text
Game-server prefixes  -> Wintun / LAGVEX
Everything else       -> normal interface
```

is consistent with existing TUN routing implementations.

**Evidence status:** DIRECTLY SUPPORTED for TUN + routing mechanics.

---

## 2.3 Avoiding tunnel routing loops is a first-class concern

Xray's TUN documentation explicitly warns that routing everything into a TUN can create an infinite loop. sing-box similarly documents `auto_detect_interface` / `default_interface` mechanisms for binding outbound connections to the physical interface to prevent routing loops under TUN. [S2][S4]

**Sources:**
- https://github.com/XTLS/Xray-core/blob/main/proxy/tun/README.md
- https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/route/index.md
- https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/inbound/tun.md

### LAGVEX requirement

The relay/tunnel transport itself must **not** accidentally enter the same route policy it is responsible for carrying.

**Recommended implementation requirement — DERIVED DESIGN:**

```text
Game destination -> tunnel
Relay endpoint   -> physical/default interface
Control endpoint -> physical/default interface
```

This is a design adaptation of the routing-loop problem documented by Xray/sing-box, not a quoted recipe from those projects.

---

# 3. Closest open-source reference implementations

## 3.1 LightSpeed — closest product pattern discovered

LightSpeed describes itself as a network optimizer for multiplayer games. Its documented model is game traffic interception followed by routing through a proxy node, with the motivation that an ISP's default path may be slower or less stable than a path through a proxy with a different backbone. [S5][S6]

Repository:
- https://github.com/ShibbityShwab/lightspeed

User guide:
- https://github.com/ShibbityShwab/lightspeed/blob/master/docs/user-guide.md

### What is useful to study

- Game traffic interception.
- Proxy-node based path substitution.
- Automatic game/server discovery concepts.
- Proxy selection concepts.
- FEC as an optional mechanism.
- Operational UX around selecting a game and a proxy.

### Important license note

The repository's own materials should be read carefully before reusing code. **Do not copy code simply because it is public on GitHub; verify the repository's current license and each dependency's license before incorporating code into LAGVEX.**

**Evidence status:** DIRECT implementation reference, not academic proof.

---

## 3.2 Glorytun — multipath UDP tunnel reference

Glorytun describes itself as a small, secure **multipath UDP tunnel**. Its README documents authenticated/encrypted traffic and multipath-related functionality. [S7][S8]

Repository:
- https://github.com/angt/glorytun
- README: https://github.com/angt/glorytun/blob/master/README.md

### What to study

- UDP tunnel structure.
- Session/authentication ideas.
- Multipath architecture.
- Secure packet processing.
- Tunnel operational concerns such as path changes and packet handling.

**Evidence status:** DIRECT implementation reference.

---

## 3.3 OpenMPTCProuter — client/VPS/multipath reference

OpenMPTCProuter is an open-source system that aggregates and encrypts multiple Internet connections and terminates the aggregated traffic at a VPS. Its README explicitly describes aggregation/failover and lists MPTCP, MLVPN, and Glorytun UDP among the underlying technologies. [S9]

Repositories:
- https://github.com/Ysurac/openmptcprouter
- https://github.com/Ysurac/openmptcprouter-vps

The VPS repository documents installation/configuration of components such as Glorytun, MPTCP, firewalling and related services. [S10]

### What to study

- Separating client-side and VPS-side roles.
- Node provisioning.
- Failover/aggregation concepts.
- Health/monitoring integration.
- VPS configuration automation.

**Do not copy the entire architecture blindly:** OpenMPTCProuter solves a broader multi-link WAN aggregation problem, not the narrower single-PC game traffic optimization problem.

**Evidence status:** DIRECT implementation reference.

---

## 3.4 UDPspeeder — FEC on lossy UDP paths

UDPspeeder is explicitly designed to improve network quality on high-latency/lossy links using Forward Error Correction. Its documentation states that FEC reduces packet loss at the cost of additional bandwidth and provides configurable FEC behavior. [S11]

Repository:
- https://github.com/wangyu-/UDPspeeder

The project documentation is particularly valuable because it exposes trade-offs and even documents cases where FEC behavior can become counterproductive due to packet grouping/loss interactions. [S11][S12]

### What to study

- FEC framing.
- Loss recovery without retransmitting original application packets.
- MTU interaction.
- Runtime parameter changes.
- Explicit tradeoff between recovery and bandwidth/queueing cost.

### LAGVEX constraint

**Do not assume FEC lowers ping.**

FEC is primarily a packet-loss mitigation mechanism. The project documentation itself describes a bandwidth cost. [S11]

**Evidence status:** DIRECT for FEC mechanics; any LAGVEX thresholds must be experimentally determined.

---

# 4. Research on latency and player experience

## 4.1 Multiplayer games: latency and jitter matter

Dick, Wellnitz, and Wolf studied player performance/perception in multiplayer games under controlled network conditions. Their work explicitly investigated latency and jitter across multiple games and examined relationships between network conditions, game score, and subjective perception. [S13]

Publication:

> Matthias Dick, Oliver Wellnitz, Lars C. Wolf, “Analysis of factors affecting players' performance and perception in multiplayer games,” Proceedings of NetGames 2005, DOI: 10.1145/1103599.1103624.

Sources:
- DBLP: https://dblp.org/rec/conf/netgames/DickWW05
- DOI: https://doi.org/10.1145/1103599.1103624
- Searchable full-text copy indexed from the publication: https://www.researchgate.net/publication/221391379_Analysis_of_factors_affecting_players%27_performance_and_perception_in_multiplayer_games

### Correct interpretation

This supports treating **latency and jitter as distinct performance variables** in multiplayer gaming.

It does **not** prove any particular LAGVEX relay algorithm or any exact latency threshold.

---

## 4.2 Real-time multiplayer QoS: low delay, low jitter and low loss

Budke et al. discuss real-time multiplayer game support and explicitly identify end-to-end delay, jitter and packet loss as important QoS concerns for real-time multiplayer games. [S14]

Source:
- https://inria.hal.science/inria-00001006/document

Again, this is background evidence supporting the metric set, not evidence that a given relay implementation will improve these metrics.

---

# 5. Internet path selection and why a relay can help

## 5.1 BGP is not latency-aware by default

Lin et al., “Latency-Aware Inter-domain Routing” (2024), analyze the problem of Internet inter-domain routing being performance-oblivious and show that circuitous routing can create latency inflation. They propose ways of incorporating abstract latency information into inter-domain routing. [S15]

Source:
- Paper: https://arxiv.org/abs/2410.13019
- DOI: https://doi.org/10.48550/arXiv.2410.13019

### Implication

A relay can be useful because it changes the Internet's entry/exit points and therefore may change the path taken toward the game server.

### Important caveat

The paper **does not prove that a consumer VPS relay will always produce lower latency**. It establishes the existence of path-selection inefficiency and studies latency-aware routing at the inter-domain routing level.

**Evidence status:** RESEARCH SUPPORT for the general path-optimization rationale, not a guarantee for LAGVEX.

---

# 6. Tail latency, jitter and real-time workloads

## 6.1 AUGUR — multipath for lower tail latency

Zhou et al. present AUGUR, a multipath transport service designed to reduce long-tail latency and video frame stall rates in mobile real-time streaming. The work identifies wireless path fluctuation and explores multipath transport as a way to improve real-time behavior while constraining cellular usage. [S16]

USENIX NSDI 2024:
- https://www.usenix.org/conference/nsdi24/presentation/zhou-yuhan
- PDF: https://www.usenix.org/system/files/nsdi24spring_prepub_zhou-yuhan.pdf
- ACM record: https://dl.acm.org/doi/10.5555/3691825.3691929

### What LAGVEX can legitimately learn

- Average/median latency is not the only metric.
- Long-tail latency can be important for real-time workloads.
- Multiple paths can be useful when path quality fluctuates.
- A multipath scheduler has to reason about path quality, not simply use every path equally.

### What it does NOT prove

AUGUR is mainly about mobile real-time streaming/cloud-gaming-like workloads; it is not a direct controlled experiment of ordinary client-rendered FPS game UDP traffic.

Do not claim “AUGUR proves multipath lowers CS2/Valorant ping.” It does not.

---

## 6.2 BLADE — Wi-Fi can be a bottleneck even when Internet/server path is good

Guo et al., “BLADE: Adaptive Wi-Fi Contention Control for Next-Generation Real-Time Communication,” NSDI 2026, report that Wi-Fi last-mile access points can remain major latency bottlenecks for real-time communication. Their large-scale measurements and evaluation identify short-term packet delivery droughts caused by Wi-Fi contention as a source of long-tail latency. [S17][S18]

Sources:
- USENIX: https://www.usenix.org/conference/nsdi26/presentation/guo-fengqian
- Paper PDF: https://www.usenix.org/system/files/nsdi26-guo-fengqian.pdf
- arXiv: https://arxiv.org/abs/2603.16119

### Critical product implication

A game relay is **not a universal cure for lag**.

Example diagnosis categories:

```text
Local Wi-Fi contention       -> relay may not fix root cause
Local queueing/bufferbloat   -> relay may not fix root cause
Bad ISP-to-game path         -> relay may help
Bad relay-to-game path       -> relay may hurt
Packet loss on chosen path   -> recovery may help
Game-server-side overload    -> network relay cannot fix it
```

The categorization above is a **DERIVED DESIGN** based on the documented bottleneck/path model. It should be validated experimentally before product claims are made.

---

# 7. Bufferbloat and queueing

## 7.1 FQ-CoDel

RFC 8290 defines Flow Queue CoDel (FQ-CoDel), a scheduler/AQM designed around flow queuing and CoDel active queue management. It is a foundational reference for understanding how queueing can inflate latency and how active queue management can reduce persistent queue delay. [S19]

Source:
- https://www.rfc-editor.org/rfc/rfc8290.html

### LAGVEX use

LAGVEX should primarily use bufferbloat ideas for **diagnosis/benchmarking** before attempting to implement a novel Windows AQM.

**DERIVED DESIGN:** run idle vs loaded RTT tests and report queue-induced latency increase.

Do not call that measurement “bufferbloat” unless the test methodology is sufficiently defined; a simple upload/download RTT increase is an indicator, not a complete standards-compliant diagnosis.

---

# 8. QUIC, UDP proxying and future transport options

## 8.1 QUIC

RFC 9000 specifies QUIC, a UDP-based secure multiplexed transport. RFC 9002 specifies QUIC loss detection and congestion control. [S20][S21]

Sources:
- RFC 9000: https://www.rfc-editor.org/rfc/rfc9000.html
- RFC 9002: https://www.rfc-editor.org/rfc/rfc9002.html

### LAGVEX recommendation

Do **not** automatically wrap every game UDP packet inside full QUIC just because QUIC is modern.

**DERIVED DESIGN:** keep a minimal native UDP tunnel as the first dataplane; consider QUIC/HTTP-based transports for future compatibility/control-plane use cases where their properties are actually useful.

The correct question is not “is QUIC faster?” but:

```text
Does the extra transport layer improve end-to-end application performance
under the actual network conditions LAGVEX targets?
```

Benchmark it.

---

## 8.2 CONNECT-UDP and HTTP proxying

RFC 9298 defines a mechanism for proxying UDP using HTTP CONNECT-UDP. RFC 9484 defines Proxying IP in HTTP. [S22][S23]

Sources:
- RFC 9298: https://www.rfc-editor.org/rfc/rfc9298.html
- RFC 9484: https://www.rfc-editor.org/rfc/rfc9484.html

### LAGVEX relevance

These are potential **future transport/interoperability options**, especially for environments where a conventional UDP tunnel is difficult to deploy.

They should not replace the initial native UDP tunnel without benchmark evidence.

---

# 9. Multipath QUIC / MPTCP

## 9.1 Multipath QUIC working area

The IETF QUIC working group maintains an active multipath QUIC working area. The repository is explicitly an in-progress Internet-Draft rather than a finished RFC. [S24]

Source:
- https://github.com/quicwg/multipath

### Critical distinction

This means:

```text
Multipath QUIC
= active standards work / implementation ecosystem
```

not:

```text
Multipath QUIC = established consumer gaming booster technique
```

Use it as research material and a future option.

---

## 9.2 MPTCP

OpenMPTCProuter provides a practical deployment reference for MPTCP-based multipath and failover. [S9][S10]

RFC 8684 specifies TCP Extensions for Multipath Operation with Multiple Addresses, i.e. MPTCP. [S25]

Source:
- RFC 8684: https://www.rfc-editor.org/rfc/rfc8684.html

### LAGVEX implication

MPTCP is more naturally applicable to TCP traffic than arbitrary game UDP datagrams. It should not be treated as a drop-in replacement for an UDP game tunnel.

---

# 10. Security / cryptography architecture

## 10.1 Do not invent crypto

Glorytun's implementation is a useful concrete reference because it documents authenticated encryption and key-management properties. [S7][S8]

Wintun itself is a virtual network interface, not the cryptographic protocol.

### LAGVEX design rule — DERIVED DESIGN

Use a standard, audited AEAD construction and a well-defined session-key handshake. Do not implement custom XOR/encryption.

The protocol should provide at minimum:

```text
Client authentication
Session key establishment
Authenticated encryption
Replay protection
Packet/session counters or equivalent nonce discipline
Key rotation policy (if long-lived sessions require it)
```

Do not claim that PSK alone constitutes a secure tunnel unless the handshake, key derivation, nonces, replay handling and authentication are correctly specified.

For the first implementation, a minimal conservative design is preferable to an elaborate custom protocol.

---

# 11. Game traffic selection

## 11.1 Static IP-range routing is simple but brittle

The current idea of routing known game-server IP ranges through Wintun is easy to reason about, but real services may use changing addresses, multiple regions, dynamic matchmaking, different ports, IPv4/IPv6 and auxiliary endpoints.

This statement is an engineering concern rather than a claim about every game; exact behavior must be measured per game.

### LAGVEX strategy — DERIVED DESIGN

Start with explicit game profiles:

```text
Game Profile
  ├── process names
  ├── known domains / discovery hints
  ├── known CIDRs (when available)
  ├── ports/protocol hints
  └── exclusions
```

Then evolve toward runtime destination discovery/learning.

### Important anti-loop requirement

Never route the tunnel's own relay/control endpoint back into the tunnel.

Xray and sing-box both document routing-loop concerns around TUN usage. [S2][S4]

---

# 12. Route measurement: what LAGVEX should measure

## 12.1 Minimum metric set

The evidence base supports considering:

```text
RTT
Jitter / latency variation
Packet loss
Tail latency
Path stability
```

Multiplayer game research directly discusses latency/jitter, while more recent real-time streaming research emphasizes tail latency and path fluctuation. [S13][S16]

### Recommended reported statistics — DERIVED DESIGN

For each candidate path, maintain a rolling window containing:

```text
median RTT
p95 RTT
p99 RTT
RTT variance / jitter metric
packet-loss rate
consecutive-loss events
probe success rate
route-switch count
```

**Do not claim that the exact set/formula is standardized.** It is a proposed measurement design for LAGVEX.

---

# 13. Relay selection

## 13.1 Why “closest VPS” is not enough

A relay route has at least two relevant segments:

```text
User → Relay
Relay → Game server
```

The complete path can also include tunnel processing and queueing overhead.

The latency-aware-routing literature supports the general idea that Internet paths can be suboptimal from a latency perspective. [S15]

### LAGVEX route scoring — DERIVED DESIGN

A first deterministic scheduler can use a weighted score:

```text
score(path) =
      w_rtt * normalized(median_RTT)
    + w_p95 * normalized(p95_RTT)
    + w_jit * normalized(jitter)
    + w_loss * normalized(packet_loss)
    + w_stab * normalized(instability)
```

This is **not a published formula being attributed to a paper**. It is a candidate LAGVEX engineering design.

### Better architecture

Do not only probe:

```text
Client → Relay
```

Also estimate/measure:

```text
Client → Relay
Relay → Game destination
```

and, where practical, the observable end-to-end session behavior.

---

# 14. Route switching and hysteresis

## Problem

If two paths are close:

```text
A = 30 ms
B = 31 ms
```

switching every few seconds can itself create instability.

### LAGVEX proposal — DERIVED DESIGN

Use path states:

```text
HEALTHY
   ↓
DEGRADED
   ↓
PROBING
   ↓
CANDIDATE
   ↓
SWITCH
```

Require either:

- a sufficiently large improvement in the new path, or
- sustained degradation of the current path,

before switching.

Add a minimum dwell time/cooldown after a switch.

**Do not invent exact thresholds until benchmark data exists.**

---

# 15. Multi-relay architecture

## Recommended architecture — DERIVED DESIGN

```text
                       ┌───────────────┐
                       │ Route Engine  │
                       └───────┬───────┘
                               │
             ┌─────────────────┼─────────────────┐
             │                 │                 │
             ▼                 ▼                 ▼
         Relay SG          Relay HK          Relay JP
             │                 │                 │
             └─────────────────┼─────────────────┘
                               │
                          Game server
```

Each relay publishes health/measurement data. The client performs active probes and selects the best current path.

### Why this is grounded

- Glorytun provides a real-world multipath UDP tunnel reference. [S7]
- OpenMPTCProuter provides a real-world multipath/failover client+VPS system. [S9][S10]
- AUGUR is academic evidence that multipath can help real-time workloads under path fluctuation, while also highlighting practical constraints. [S16]

### What is not established

There is no source in this research set that proves the exact LAGVEX multi-VPS scheduler described above will improve FPS gameplay. It must be benchmarked.

---

# 16. FEC / packet-loss recovery

## Evidence

UDPspeeder directly documents FEC as a mechanism to reduce packet loss on lossy links, with bandwidth cost and MTU/grouping trade-offs. [S11][S12]

## LAGVEX design recommendation — DERIVED DESIGN

Treat FEC as an adaptive recovery layer, not a permanent “ping booster” switch.

Conceptual policy:

```text
loss negligible
    → FEC disabled/minimal

loss elevated
    → light recovery

loss persistent/high
    → stronger recovery, subject to bandwidth and queueing limits
```

The actual thresholds must be determined from experiments.

### Important

For small real-time game packets, a bad FEC design can introduce delay or excess bandwidth. Do not assume redundancy is free.

---

# 17. Packet pacing / queue control

## Evidence

RFC 8290 establishes FQ-CoDel as an AQM/flow-queueing reference for controlling queueing behavior. [S19]

BLADE shows, in a real-time communication context, that Wi-Fi contention can create long-tail packet delivery problems that are not solved by simply assuming the Internet backbone is the bottleneck. [S17][S18]

## LAGVEX design recommendation — DERIVED DESIGN

Investigate packet pacing only after measuring whether LAGVEX itself creates burst/queueing problems.

Do not add a home-grown AQM in V1.

Start with:

```text
socket buffer observability
packet timestamps
inter-packet timing measurements
send-rate telemetry
queue-length telemetry where available
```

Then benchmark whether pacing changes:

```text
median RTT
p95 RTT
p99 RTT
loss
CPU
bandwidth
```

---

# 18. Diagnostics: detect when LAGVEX should NOT be used

## This should become a first-class feature

A network accelerator should be allowed to answer:

> “Your relay is not the bottleneck. Do not use it.”

### Proposed decision tree — DERIVED DESIGN

```text
                    Start
                      |
                Measure direct
                      |
              ┌───────┴───────┐
              |               |
       Stable direct      Unstable direct
              |               |
          Compare         Diagnose local
           relays          vs upstream
              |               |
        ┌─────┴─────┐         |
        |           |         |
   Relay improves  Relay worse |
        |           |         |
      BOOST        DIRECT    report cause
```

### Possible diagnostic output

```text
DIRECT
  Median: 31 ms
  P95:    44 ms
  Loss:   0.2%

LAGVEX / HK
  Median: 34 ms
  P95:    35 ms
  Loss:   0.0%

Recommendation: BOOST
Reason: Higher median latency, but substantially better tail stability.
```

That example is synthetic. It is **not measured data**.

---

# 19. Direct vs relay benchmarking

## Required benchmark modes

Every optimization claim should have at least:

```text
A. DIRECT
B. RELAY-ONE
C. RELAY-TWO / alternate path
```

and ideally:

```text
D. BAD-network simulation
E. packet-loss test
F. queueing/bufferbloat test
G. route-switch test
```

### Metrics to record

```text
median RTT
p95 RTT
p99 RTT
jitter
packet loss
CPU utilization
memory
bandwidth overhead
connection setup time
route-switch interruption
```

### Why p95/p99?

AUGUR's research focus is specifically on long-tail latency in real-time streaming, which motivates looking beyond mean latency for real-time workloads. [S16]

Again, the exact LAGVEX benchmark protocol is a **DERIVED DESIGN**.

---

# 20. Transport protocol: recommended V1

## V1 proposal — DERIVED DESIGN

```text
Game UDP packet
    ↓
LAGVEX packet header
    ↓
AEAD protected datagram
    ↓
UDP socket
    ↓
VPS relay
    ↓
AEAD verify/decrypt
    ↓
Game server
```

### Keep V1 simple

Avoid initially adding:

```text
QUIC
FEC
multipath
compression
complex stream multiplexing
```

unless a measured problem requires them.

This is an engineering recommendation, not a standards requirement.

---

# 21. Control plane vs data plane

## Recommended separation — DERIVED DESIGN

```text
                CONTROL PLANE
        ┌───────────────────────────┐
        │ relay registry            │
        │ node health               │
        │ authentication metadata   │
        │ game profiles             │
        │ configuration              │
        │ telemetry summaries        │
        └────────────┬──────────────┘
                     │
                     ▼
                DATA PLANE
        ┌───────────────────────────┐
        │ packet receive             │
        │ decrypt/authenticate       │
        │ lookup session             │
        │ forward                    │
        └───────────────────────────┘
```

The relay dataplane should stay extremely small and predictable.

**Why:** minimizing work in the packet-forwarding hot path reduces the number of things that can add CPU/queueing overhead. This is an engineering principle, not a direct quote from a specific paper in this document.

---

# 22. Windows implementation references

## Wintun

Official:
- https://github.com/WireGuard/wintun

Important facts:
- Layer 3 TUN adapter.
- Userspace-oriented.
- Official deployment uses `wintun.dll`. [S1]

## Xray Windows TUN implementation

Official source:
- https://github.com/XTLS/Xray-core/blob/main/proxy/tun/tun_windows.go

Useful implementation references:
- Wintun adapter creation.
- Windows IP configuration.
- route handling.
- adapter/session lifecycle. [S3]

## sing-box TUN/routing

Official:
- https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/inbound/tun.md
- https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/route/index.md

Useful references:
- TUN routing.
- interface selection.
- routing-loop prevention concepts. [S4]

---

# 23. Suggested repository architecture for LAGVEX

**This entire section is a DERIVED DESIGN proposal, not a copy of an existing project architecture.**

```text
lagvex/
│
├── client/
│   ├── wintun/
│   ├── routing/
│   ├── classifier/
│   ├── tunnel/
│   ├── probing/
│   └── platform/windows/
│
├── relay/
│   ├── transport/
│   ├── crypto/
│   ├── session/
│   ├── forwarding/
│   └── metrics/
│
├── scheduler/
│   ├── metrics/
│   ├── scoring/
│   ├── hysteresis/
│   └── failover/
│
├── diagnostics/
│   ├── latency/
│   ├── jitter/
│   ├── loss/
│   ├── queueing/
│   └── report/
│
├── games/
│   ├── profiles/
│   └── discovery/
│
├── protocol/
│   ├── wire-format/
│   └── versioning/
│
├── deploy/
│   ├── vps/
│   └── systemd/
│
├── benchmark/
│   ├── direct/
│   ├── relay/
│   ├── loss/
│   ├── jitter/
│   └── failover/
│
└── docs/
    ├── architecture.md
    ├── protocol.md
    ├── security.md
    └── benchmarking.md
```

---

# 24. Suggested implementation phases

## Phase 1 — reliable tunnel

```text
Wintun
→ selective route
→ UDP tunnel
→ authenticated encryption
→ VPS relay
→ game destination
```

Acceptance criteria:

- Packets can traverse the tunnel reliably.
- Relay does not route itself back into the tunnel.
- Normal non-game traffic remains direct.
- Tunnel teardown restores routing cleanly.
- MTU behavior is explicitly tested.

Reference foundations: Wintun [S1], Xray TUN/routing [S2][S3], Glorytun [S7].

---

## Phase 2 — observability

Implement:

```text
RTT
jitter
loss
median
p95
p99
packet counters
bytes
CPU
```

Acceptance criteria:

- Same test can be reproduced for direct and relay paths.
- Data can be exported to JSON/CSV.
- No performance claims are made from a single ping sample.

Research motivation: multiplayer latency/jitter [S13][S14], tail latency [S16].

---

## Phase 3 — relay selection

Implement:

```text
relay registry
active probes
path score
best-relay selection
```

Acceptance criteria:

- Relay A/B can be objectively compared.
- Selection uses more than a single instantaneous ping.
- Selection decisions are logged with measurements.

Research motivation: latency-aware routing [S15].

---

## Phase 4 — stability and failover

Implement:

```text
health states
hysteresis
cooldown
backup relay
automatic failover
```

Reference motivation: Glorytun [S7], OpenMPTCProuter [S9][S10], multipath research [S16].

---

## Phase 5 — diagnostics

Add:

```text
Direct vs relay report
local queueing indicator
Wi-Fi/local-path warning
route quality diagnosis
```

Research motivation: BLADE [S17][S18], FQ-CoDel/bufferbloat background [S19].

---

## Phase 6 — adaptive recovery

Possible features:

```text
adaptive FEC
loss-aware recovery
packet pacing experiments
```

Primary reference for FEC: UDPspeeder [S11][S12].

Do not enable these globally without benchmarks.

---

## Phase 7 — advanced multipath

Research options:

```text
multipath relay set
path scheduler
packet duplication for selected conditions
recovery/reinjection
Multipath QUIC
MPTCP-related ideas
```

References: AUGUR [S16], Glorytun [S7], OpenMPTCProuter [S9], IETF Multipath QUIC work [S24], RFC 8684 [S25].

---

# 25. What NOT to build first

## 25.1 Registry “ping tweaks”

Do not make registry tweaks the foundation of LAGVEX.

They do not solve the main problem LAGVEX is designed around: selecting a potentially better network path for game traffic.

If a tweak is considered later, it needs independent evidence and a benchmark.

---

## 25.2 “Change DNS = lower ping” as a core feature

DNS can influence name resolution and therefore potentially affect endpoint selection in some systems, but DNS replacement is not equivalent to optimizing the subsequent packet route.

Do not market DNS switching as the central latency mechanism.

---

## 25.3 “Always use the closest relay”

Not evidence-based.

The relevant question is whether the **end-to-end path** through that relay performs better.

This follows the path-optimization rationale in [S15] and is a derived engineering requirement.

---

## 25.4 “Always enable FEC”

Not justified. UDPspeeder explicitly documents bandwidth and FEC/packet grouping trade-offs. [S11][S12]

---

## 25.5 “Always use multiple relays simultaneously”

Not justified for ordinary FPS traffic without measurements.

Multipath research shows potential benefits in specific real-time workloads, but ordinary multiplayer FPS traffic is not identical to mobile/cloud video streaming. [S16]

---

# 26. Important scope distinction: FPS multiplayer vs cloud gaming

This distinction must be preserved in all future documentation.

### Ordinary client-rendered multiplayer FPS

Conceptually:

```text
Game client
   ↕
small interactive packets
   ↕
authoritative game server
```

### Cloud gaming

Conceptually:

```text
input
  ↓
remote game instance
  ↓
encoded video
  ↓
client
```

Papers such as AUGUR, BLADE, and GeForce NOW network studies often address cloud/real-time streaming. [S16][S17][S26]

They are still valuable for:

- tail latency,
- path instability,
- wireless bottlenecks,
- measurement methodology,

but they should **not** be presented as direct experimental proof for competitive FPS packet transport.

---

# 27. Benchmark methodology

## 27.1 Basic A/B test

For every game/profile:

```text
Test 1: DIRECT
Test 2: LAGVEX relay
```

Run multiple trials, not a single 10-second ping.

Collect:

```text
N samples
median
mean
p95
p99
jitter
loss
```

---

## 27.2 Controlled impairments

Use a controlled lab environment to emulate:

```text
+ latency
+ jitter
+ packet loss
+ bandwidth constraint
+ queueing
```

This helps determine whether adaptive mechanisms activate in the situations they are designed to handle.

A paper/research environment cited in the gaming literature uses Linux `netem` to introduce network impairments for game studies. See the published player-perception study indexed at [S13] and related experimental material.

---

## 27.3 Relay comparison experiment

For every region:

```text
Direct
Relay A
Relay B
Relay C
```

Record:

```text
client → relay RTT
relay → destination RTT (where measurable)
end-to-end observed behavior
p95/p99
loss
stability
```

Then determine whether the relay improved the actual path.

---

## 27.4 Route-switch experiment

Force a route degradation event:

```text
Current relay
     ↓
quality degrades
     ↓
probe alternatives
     ↓
select better route
     ↓
switch
```

Measure:

```text
switch duration
packet loss during switch
RTT before/after
post-switch stability
```

---

# 28. Telemetry schema proposal

**DERIVED DESIGN.**

```json
{
  "timestamp": 0,
  "game": "example",
  "path": "relay-hk-01",
  "rtt_ms": 0,
  "jitter_ms": 0,
  "p50_ms": 0,
  "p95_ms": 0,
  "p99_ms": 0,
  "packet_loss": 0.0,
  "probe_success": 1.0,
  "route_state": "HEALTHY",
  "bytes_tx": 0,
  "bytes_rx": 0,
  "cpu_percent": 0.0
}
```

Do not collect user-identifying information by default.

---

# 29. Relay protocol design requirements

**DERIVED DESIGN.**

Minimum conceptual packet/session model:

```text
Handshake
  ↓
Authenticated session
  ↓
Encrypted datagrams
  ↓
Sequence / replay protection
  ↓
Forwarding
```

### Avoid

```text
custom cipher
XOR-only obfuscation
unauthenticated packet encryption
unbounded receive buffers
implicit trust of any client
```

### Relay should enforce

```text
session authentication
rate/bandwidth sanity limits
packet-size sanity
replay protection
connection cleanup
idle timeout
```

These are security engineering requirements; they are not being attributed to a single cited paper here.

---

# 30. MTU / fragmentation research requirement

Tunnel encapsulation changes the effective packet size available to the inner traffic. UDPspeeder's documentation demonstrates why MTU and packet splitting/merging become important when doing FEC/tunnel processing. [S12]

### LAGVEX requirement — DERIVED DESIGN

V1 must explicitly test:

```text
normal-size packets
near-MTU packets
IPv4
IPv6
fragmentation behavior
encapsulation overhead
```

Do not simply assume `MTU = 1500` is always safe for an encapsulated protocol.

---

# 31. Future research directions

## A. Adaptive path scheduler

Use rolling path-quality estimates and select based on latency distribution + stability.

Evidence basis: latency-aware routing [S15], AUGUR [S16].

Status for LAGVEX: **recommended research direction; exact algorithm is DERIVED DESIGN.**

---

## B. Multipath / redundant paths

Investigate whether a small amount of redundancy can reduce the impact of short losses without causing excessive bandwidth or packet reordering.

Evidence basis: Glorytun [S7], AUGUR [S16], Multipath QUIC work [S24].

Status: **experimental.**

---

## C. Adaptive FEC

Enable recovery only when measured loss warrants it.

Evidence basis: UDPspeeder [S11][S12].

Status: **experimental.**

---

## D. Local bottleneck diagnosis

Detect situations where the user's own Wi-Fi/queueing is the dominant issue.

Evidence basis: BLADE [S17][S18], FQ-CoDel [S19].

Status: **high-value feature; exact classifier is DERIVED DESIGN.**

---

## E. More robust game destination discovery

Move from static IP lists toward profile + runtime discovery.

Evidence basis for TUN/process/routing architectures: Xray [S2][S3], sing-box [S4], LightSpeed [S5][S6].

Status: **engineering roadmap.**

---

## F. HTTP/3 / MASQUE-compatible transport

Investigate CONNECT-UDP / CONNECT-IP for difficult network environments.

Evidence basis: RFC 9298 [S22], RFC 9484 [S23].

Status: **long-term / compatibility research.**

---

# 32. Proposed LAGVEX architecture, consolidated

**DERIVED DESIGN — synthesized from the cited evidence and not an existing published LAGVEX architecture.**

```text
                         ┌─────────────────────────┐
                         │      LAGVEX CLIENT      │
                         └────────────┬────────────┘
                                      │
                             Game detection
                                      │
                                      ▼
                         ┌─────────────────────────┐
                         │ Destination / Profile   │
                         │ manager                 │
                         └────────────┬────────────┘
                                      │
                                      ▼
                         ┌─────────────────────────┐
                         │     Windows Wintun      │
                         └────────────┬────────────┘
                                      │
                                      ▼
                         ┌─────────────────────────┐
                         │ Selective packet router │
                         └────────────┬────────────┘
                                      │
                             game traffic only
                                      │
                                      ▼
                         ┌─────────────────────────┐
                         │     Tunnel engine       │
                         │     AEAD / UDP V1       │
                         └────────────┬────────────┘
                                      │
                           ┌──────────┼──────────┐
                           │          │          │
                           ▼          ▼          ▼
                      Relay SG   Relay HK   Relay JP
                           │          │          │
                           └──────────┼──────────┘
                                      │
                                      ▼
                                Game server

                  ┌────────────────────────────────────┐
                  │         MEASUREMENT LOOP            │
                  │ RTT / p95 / p99 / jitter / loss    │
                  └────────────────┬───────────────────┘
                                   │
                                   ▼
                          Route scoring engine
                                   │
                                   ▼
                          Hysteresis / failover
```

---

# 33. Agent instructions

The implementation agent should follow these rules:

### Rule 1 — evidence discipline

Never write comments/docs such as:

> “This technique is scientifically proven to lower ping.”

unless a cited primary source actually supports that exact claim and workload.

Prefer:

> “This mechanism is inspired by X; LAGVEX performance must be benchmarked independently.”

### Rule 2 — preserve the distinction between facts and proposals

Use labels:

```text
[SOURCE]
[DERIVED DESIGN]
[EXPERIMENTAL]
[UNVERIFIED]
```

### Rule 3 — no magic optimization values

Do not hard-code claims such as:

```text
switch after 20 ms
FEC above 1%
use relay if 5 ms lower
```

without benchmark justification.

### Rule 4 — benchmark every optimization

Every new mechanism should have a toggle and an A/B benchmark.

### Rule 5 — keep V1 minimal

First establish a correct, low-overhead tunnel and measurement system. Then add intelligence.

### Rule 6 — never sacrifice safety for latency

Do not weaken authentication/replay protection to gain an unmeasured performance benefit.

### Rule 7 — do not hide failure

If a relay is worse than direct, LAGVEX should be able to report that and stay direct.

---

# 34. Recommended V1/V2/V3 feature matrix

| Feature | V1 | V2 | V3+ | Evidence / status |
|---|---:|---:|---:|---|
| Wintun | YES | | | Official Wintun [S1] |
| Selective routing | YES | | | Xray/sing-box implementation references [S2][S3][S4] |
| UDP relay | YES | | | LightSpeed/Glorytun-style reference [S5][S7] |
| Authenticated encryption | YES | | | Security requirement; Glorytun as implementation reference [S7] |
| Active probes | YES | | | Derived design |
| RTT + loss | YES | | | Gaming/real-time research [S13][S14] |
| p95/p99 | YES | | | Tail-latency motivation [S16] |
| Automatic relay selection | | YES | | Derived design; latency-aware routing motivation [S15] |
| Relay failover | | YES | | Glorytun/OpenMPTCProuter references [S7][S9] |
| Local bottleneck diagnosis | | YES | | BLADE/FQ-CoDel motivation [S17][S19] |
| Adaptive FEC | | | YES | UDPspeeder [S11][S12] |
| Multipath | | | YES | Glorytun/OpenMPTCProuter/AUGUR [S7][S9][S16] |
| Packet pacing | | | YES | Experimental; queueing research context [S19] |
| Multipath QUIC | | | YES | IETF work [S24] |
| CONNECT-UDP | | | YES | RFC 9298 [S22] |
| CONNECT-IP | | | YES | RFC 9484 [S23] |

---

# 35. Source register

## S1 — Wintun

WireGuard, “Wintun Network Adapter.” Official GitHub repository.

https://github.com/WireGuard/wintun

Official website:
https://www.wintun.net/

Use for: Layer-3 TUN/Windows architecture and deployment model.

---

## S2 — Xray TUN documentation

XTLS/Xray-core, `proxy/tun/README.md`.

https://github.com/XTLS/Xray-core/blob/main/proxy/tun/README.md

Use for: Windows Wintun TUN support, route configuration, route-loop warnings.

---

## S3 — Xray Windows TUN implementation

XTLS/Xray-core, `proxy/tun/tun_windows.go`.

https://github.com/XTLS/Xray-core/blob/main/proxy/tun/tun_windows.go

Use for: concrete Windows Wintun and route-management implementation details.

---

## S4 — sing-box TUN/routing documentation

SagerNet/sing-box.

TUN:
https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/inbound/tun.md

Routing:
https://github.com/SagerNet/sing-box/blob/testing/docs/configuration/route/index.md

Use for: TUN routing, interface selection, routing-loop avoidance concepts.

---

## S5 — LightSpeed repository

ShibbityShwab, “LightSpeed.”

https://github.com/ShibbityShwab/lightspeed

Use for: open-source game traffic accelerator/proxy pattern.

---

## S6 — LightSpeed user guide

https://github.com/ShibbityShwab/lightspeed/blob/master/docs/user-guide.md

Use for: documented game traffic interception, proxy routing and FEC workflow.

---

## S7 — Glorytun

angt, “Glorytun: Multipath UDP tunnel.”

https://github.com/angt/glorytun

Use for: secure UDP tunnel and multipath implementation reference.

---

## S8 — Glorytun README

https://github.com/angt/glorytun/blob/master/README.md

Use for: documented cryptographic/session features.

---

## S9 — OpenMPTCProuter

Ysurac, “OpenMPTCProuter.”

https://github.com/Ysurac/openmptcprouter

Use for: client/VPS, multipath, aggregation and failover architecture.

---

## S10 — OpenMPTCProuter VPS

https://github.com/Ysurac/openmptcprouter-vps

Use for: VPS-side provisioning and tunnel-related components.

---

## S11 — UDPspeeder

wangyu-, “UDPspeeder.”

https://github.com/wangyu-/UDPspeeder

Use for: FEC on lossy UDP links, bandwidth tradeoffs, runtime settings.

---

## S12 — UDPspeeder mode/MTU documentation

https://github.com/wangyu-/UDPspeeder/wiki/mode-0-vs-mode-1

https://github.com/wangyu-/UDPspeeder/wiki/About-Large-Packet%27s-Passthrough

Use for: FEC modes, packet splitting, MTU considerations.

---

## S13 — Dick, Wellnitz, Wolf (2005)

“Analysis of factors affecting players' performance and perception in multiplayer games.” NETGAMES 2005.

DOI:
https://doi.org/10.1145/1103599.1103624

DBLP:
https://dblp.org/rec/conf/netgames/DickWW05

Use for: empirical study of latency/jitter effects in multiplayer games.

---

## S14 — Budke et al. (2006)

“Real-Time Multiplayer Game Support Using QoS Mechanisms in Mobile Ad Hoc Networks.”

Full text:
https://inria.hal.science/inria-00001006/document

Use for: delay/jitter/loss as QoS concerns in real-time multiplayer games.

---

## S15 — Lin et al. (2024)

“Latency-Aware Inter-domain Routing.”

Authors: Shihan Lin, Yi Zhou, Xiao Zhang, Todd Arnold, Ramesh Govindan, Xiaowei Yang.

https://arxiv.org/abs/2410.13019

DOI:
https://doi.org/10.48550/arXiv.2410.13019

Use for: inter-domain routing inefficiency and latency-aware path selection motivation.

---

## S16 — AUGUR, NSDI 2024

Yuhan Zhou et al., “AUGUR: Practical Mobile Multipath Transport Service for Low Tail Latency in Real-Time Streaming.”

USENIX:
https://www.usenix.org/conference/nsdi24/presentation/zhou-yuhan

PDF:
https://www.usenix.org/system/files/nsdi24spring_prepub_zhou-yuhan.pdf

ACM:
https://dl.acm.org/doi/10.5555/3691825.3691929

Use for: tail latency, multipath, real-time workload measurement.

---

## S17 — BLADE, NSDI 2026

Fengqian Guo et al., “BLADE: Adaptive Wi-Fi Contention Control for Next-Generation Real-Time Communication.”

USENIX:
https://www.usenix.org/conference/nsdi26/presentation/guo-fengqian

PDF:
https://www.usenix.org/system/files/nsdi26-guo-fengqian.pdf

Use for: Wi-Fi last-mile contention and long-tail latency in real-time communication.

---

## S18 — BLADE arXiv

https://arxiv.org/abs/2603.16119

Use for: freely accessible preprint and abstract/details.

---

## S19 — RFC 8290, FQ-CoDel

Hoeiland-Joergensen et al., “The Flow Queue CoDel Packet Scheduler and Active Queue Management Algorithm.”

https://www.rfc-editor.org/rfc/rfc8290.html

Use for: queue management/bufferbloat background.

---

## S20 — RFC 9000, QUIC

“IETF QUIC: A UDP-Based Multiplexed and Secure Transport.”

https://www.rfc-editor.org/rfc/rfc9000.html

Use for: QUIC protocol semantics.

---

## S21 — RFC 9002, QUIC loss detection/congestion control

https://www.rfc-editor.org/rfc/rfc9002.html

Use for: QUIC RTT/loss/congestion-control semantics.

---

## S22 — RFC 9298, CONNECT-UDP

“Proxying UDP in HTTP.”

https://www.rfc-editor.org/rfc/rfc9298.html

Use for: HTTP-based UDP proxying option.

---

## S23 — RFC 9484, CONNECT-IP / Proxying IP in HTTP

https://www.rfc-editor.org/rfc/rfc9484.html

Use for: HTTP-based IP proxying option.

---

## S24 — IETF Multipath QUIC working area

https://github.com/quicwg/multipath

Use for: current Multipath QUIC standardization work. Treat as work in progress, not a final RFC unless independently verified.

---

## S25 — RFC 8684, MPTCP

“TCP Extensions for Multipath Operation with Multiple Addresses.”

https://www.rfc-editor.org/rfc/rfc8684.html

Use for: MPTCP protocol reference.

---

## S26 — Network Anatomy and Real-Time Measurement of NVIDIA GeForce NOW Cloud Gaming

Minzhao Lyu, Sharat Chandra Madanapalli, Arun Vishwanath, Vijay Sivaraman, PAM 2024.

arXiv:
https://arxiv.org/abs/2401.06366

Author/publication page:
https://minzhaolyu.github.io/publications/

DOI:
https://doi.org/10.1007/978-3-031-56249-5_3

Use for: real-time measurement methodology and contextual network factors in cloud gaming.

---

# 36. Final implementation doctrine

The highest-confidence architecture for the first serious LAGVEX release is:

```text
Selective game traffic
        ↓
Windows Wintun
        ↓
Minimal authenticated UDP tunnel
        ↓
Self-hosted Go relay
        ↓
Game server
```

Then build intelligence around it:

```text
Measure
  ↓
Compare direct vs relay
  ↓
Track tail latency / jitter / loss
  ↓
Choose path
  ↓
Monitor continuously
  ↓
Fail over when justified
```

Then research advanced mechanisms:

```text
FEC
Multipath
Pacing
Adaptive schedulers
QUIC / MASQUE
```

Only add them when an experiment demonstrates a real benefit for the target workload.

The core product principle should be:

> **LAGVEX does not promise to magically reduce ping. It measures whether an alternate path is better, uses it when evidence says it is better, and gets out of the way when it is not.**

That wording is an engineering/product philosophy, not a quoted research result.

---

# 24. Production Engineering Audit & Concrete Priority Roadmap (P0 — P4)

A rigorous technical review of the current implementation against the research baseline has identified critical priorities required to transition from an MVP to a secure, production-grade system:

## 24.1 Priority P0: Cryptographic Hardening & Anti-Replay
1. **Data-Plane AEAD**: Eliminate plaintext payload vulnerability by wrapping each game packet inside **ChaCha20-Poly1305** using session keys derived during the HMAC-SHA256 handshake.
2. **Server-Side Nonce Cache**: Replace naive $\pm 120$s timestamp validation with an active sliding-window nonce cache to strictly reject replayed handshake attempts.
3. **Authenticated Dynamic Roaming**: Require successful AEAD tag verification before calling `sess.UpdateRemote(remote)` to block IP-spoofing session takeovers.
4. **Honest Claims**: Refrain from using absolute claims ("100% Anti-Cheat Safe" or "zero jitter") in user-facing documentation.

## 24.2 Priority P1: Native OS Routing & State Invariants
1. **Windows IP Helper APIs**: Migrate from CLI string scraping (`route print`, `netsh`) to native Win32 `iphlpapi.dll` APIs (`GetBestRoute2`, `GetAdaptersAddresses`, `CreateIpForwardEntry2`) to avoid localization failures.
2. **Atomic Session Cleanup**: When a client reconnects, terminate any pre-existing session for that `ClientID` prior to allocating a new IP address from the pool.

## 24.3 Priority P2: True Telemetry & Tail-Latency HUD
1. **Direct vs. Relay Telemetry**: Provide comparative RTT measurements to dynamically recommend whether boosting is advantageous.
2. **Tail Latency (P95/P99) & Jitter**: Render rolling histograms of latency variation on the client HUD.

## 24.4 Priority P3 & P4: Multi-Relay Schedulers, Hysteresis & Adaptive FEC
1. **Hysteresis Anti-Flapping**: Prevent route thrashing across multiple relay nodes.
2. **Adaptive FEC**: Inject parity packets exclusively when undersea link loss exceeds acceptable thresholds.

For the comprehensive operational specification, refer to [docs/ROADMAP.md](ROADMAP.md).

---

[⬅️ Back to README](../README.md) • [🤝 Contributor Guide](../CONTRIBUTING.md) • [🗺️ Roadmap](ROADMAP.md) • [📐 Architecture](ARCHITECTURE.md) • [⚖️ Legal Disclaimer](DISCLAIMER.md)
