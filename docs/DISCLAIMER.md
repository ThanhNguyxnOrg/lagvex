# ⚖️ Legal Disclaimer & Ban Liability Waiver 🛡️

> [!CAUTION]
> **CRITICAL LEGAL NOTICE: PLEASE READ THIS DOCUMENT IN ITS ENTIRETY PRIOR TO INSTALLING, COMPILING, OR EXECUTING LAGVEX.**  
> 🔗 [Back to Project README.md](../README.md)

---

## 1. 🔬 Academic, Research & Educational Scope

**Lagvex** is an open-source technical research demonstration exploring:
- 🔌 **Layer-3 Virtual Network Interfaces**: Architecture and performance of user-space packet rings (`wintun` on Windows, `utun` on macOS, and `/dev/net/tun` on Linux).
- 🛣️ **Dynamic Routing Table Management**: Safe application and deletion of `store=active` (RAM-only) transient routes to minimize blast radius on client machines.
- ⚡ **Lightweight UDP Encapsulation**: Performance characteristics of minimal fixed-header binary tunneling with HMAC-SHA256 authenticated handshakes.
- 🐧 **Zero-Copy Kernel NAT**: Benchmarking Linux kernel `iptables` MASQUERADE and TCP MSS clamping over virtual interfaces.

This codebase is provided as free reference material for network engineers, students, and systems researchers.

---

## 2. 🚫 Zero Game Tampering & Anti-Cheat Ethics

Lagvex is strictly a network routing tool. To preserve competitive integrity:
- ❌ **No Virtual Memory Access**: Lagvex never invokes `OpenProcess`, `ReadProcessMemory`, or `WriteProcessMemory`. It cannot read or alter game state.
- ❌ **No Code / DLL Injection**: Lagvex does not inject `.dll` files, load libraries into game addresses, or install API detours/hooks.
- ❌ **No Socket Hooking (WinDivert / LSP / WFP)**: Unlike traditional tools, Lagvex never sits inside the game's internal socket layer.
- ❌ **No Gameplay Exploits**: Lagvex does not alter packet contents, forge timestamps, duplicate inputs, or provide any unfair gameplay advantage.

---

## 3. 🛑 COMPLETE DISCLAIMER OF LIABILITY REGARDING ACCOUNT BANS ⚠️

> [!WARNING]
> ### 🚨 READ CAREFULLY: NO RESPONSIBILITY FOR ACCOUNT SANCTIONS 🚨
> 
> Game publishers and anti-cheat developers—including **Riot Games, Valve Corporation, Krafton, Electronic Arts (EA), Activision Blizzard, Ubisoft, Tencent, and BattlEye Innovations**—maintain proprietary, closed-source security solutions (such as **Riot Vanguard, Valve Anti-Cheat, BattlEye, Easy Anti-Cheat, RICOCHET, and ACE**).
> 
> 1. **Proprietary Anti-Cheat Autonomy**: Anti-cheat systems use automated behavioral heuristics and telemetry. While Lagvex operates purely at the OS network layer (identical to an enterprise VPN), the developers of Lagvex **CANNOT guarantee** that a third-party anti-cheat will not flag network proxies or virtual network adapters.
> 2. **Terms of Service (ToS) Discretion**: Game publishers retain the unilateral legal right to terminate, suspend, or restrict accounts for any reason, including the use of third-party network routing software.
> 3. **EXPRESS WAIVER OF LIABILITY**:  
>    **UNDER NO LEGAL THEORY (WHETHER IN CONTRACT, TORT, NEGLIGENCE, OR OTHERWISE) SHALL THE AUTHORS, REPOSITORY OWNERS, MAINTAINERS, OR CONTRIBUTORS OF LAGVEX BE LIABLE FOR ANY:**
>    - ⛔ **Permanent or Temporary Account Bans / Suspensions**
>    - 💻 **Hardware ID (HWID) Bans or System Blacklisting**
>    - 🎖️ **Loss of Competitive Ranks, Match History, or Leaderboard Standing**
>    - 💸 **Loss of In-Game Currency, Skins, Battle Passes, or Digital Purchases**
>    - 📉 **Damages, System Instability, or Network Downtime**
> 4. **User Assumption of Risk**: By using Lagvex, you acknowledge that you have read and understood this notice and that **you choose to use this software entirely at your own individual risk and discretion**.

---

## 4. 📜 License & "AS IS" Warranty Clause

As stipulated in the **[MIT License](../LICENSE)**:

> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---

## 5. 🏷️ Nominative Fair Use & Intellectual Property Notice ⚖️

All game titles, publisher/developer names, trade dress, and associated insignias (including *Valorant*, *Counter-Strike 2*, *PUBG*, *Apex Legends*, *League of Legends*, *The Finals*, *Call of Duty*, *Delta Force*, *Overwatch 2*, *Rainbow Six Siege*, *Dota 2*, and their respective marks belonging to Valve, Riot Games, Electronic Arts, Krafton, Activision, Ubisoft, and TiMi Studio Group) are trademarks or registered trademarks of their respective legal copyright holders.

Lagvex is an independent open-source network protocol research tool and kernel latency accelerator. It is **not** endorsed by, affiliated with, sponsored by, or associated with any game developer or publisher. All vector marks in this repository and web UI are reproduced strictly under **Nominative Fair Use** solely to identify game executables and destination routing clusters for end-user network optimization.

---

🔗 **Return to Main Documentation**:
- 🏠 [**Project README**](../README.md)
- 📐 [**Architecture Overview**](ARCHITECTURE.md)
- 📡 [**Protocol Specification**](PROTOCOL.md)
- 🚀 [**Deployment Guide**](DEPLOYMENT.md)
- 🌐 [**Game Profiles & CIDRs**](PROFILES.md)
