# ⚡ Lagvex

> **Open-Source Gaming Latency Reducer & Ping Booster**  
> *Chơi game ping thấp, ổn định đường truyền, chống nghẽn mạng ISP bằng kỹ thuật Route-based Split-Tunneling an toàn tuyệt đối.*

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux-blue)](https://github.com/lagvex/lagvex)
[![Anti-Cheat](https://img.shields.io/badge/Anti--Cheat-100%25%20Safe%20(No%20Injection)-brightgreen)](https://github.com/lagvex/lagvex)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## 🎯 Tại sao Lagvex vượt trội?

Khi chơi game online (như Valorant, CS2, PUBG, Apex Legends), đường truyền mặc định của nhà mạng (ISP) tại Việt Nam thường đi qua các tuyến cáp quang bị nghẽn hoặc định tuyến vòng vèo sang Hong Kong, Nhật Bản rồi mới tới Singapore, gây ra hiện tượng **ping cao, giật lag, mất gói (packet loss) và choke**.

Các giải pháp truyền thống:
- **VPN thông thường**: Đẩy toàn bộ mạng (Web, YouTube, Discord, Windows Update) qua VPN, làm nghẽn băng thông, Discord bị đổi vùng và tốn tài nguyên.
- **Phần mềm hook game (DLL Injection / Memory read / WinDivert)**: Rất dễ bị các hệ thống Anti-Cheat khắt khe (Riot Vanguard, BattlEye, EasyAntiCheat, VAC) quét và cấm tài khoản vĩnh viễn (False-positive Ban).

**Lagvex giải quyết triệt để vấn đề này**:
1. **100% An toàn với Anti-Cheat**: Không đọc bộ nhớ game, không DLL injection, không can thiệp socket. Chỉ sử dụng card mạng ảo **WinTun** (từ dự án WireGuard) và sửa **bảng định tuyến Windows (Windows Routing Table)**.
2. **Split-Tunneling theo dải IP Game Server**: Chỉ lưu lượng hướng tới dải IP máy chủ game mới đi qua tunnel. Discord, trình duyệt, download vẫn đi trực tiếp qua mạng nhà với tốc độ tối đa.
3. **Relay Server tự host siêu nhẹ (Go)**: Nhận gói đã bọc 9-byte header siêu nhỏ gọn qua UDP, đẩy thẳng vào Linux Kernel TUN device để kernel tự làm NAT MASQUERADE và MSS clamping. Độ trễ bổ sung gần như bằng 0ms.
4. **Hỗ trợ đa game hot hiện nay**: Valorant, CS2, PUBG, Apex Legends, The Finals, Call of Duty, Delta Force, Overwatch 2, Rainbow Six Siege, League of Legends, Dota 2.

---

## 🎮 Danh mục Game hỗ trợ sẵn (Out-of-the-Box)

| Tựa Game | Thể loại | Cụm Server (Region) | Anti-Cheat Tương thích |
|---|---|---|---|
| **Valorant** | Tactical FPS | Singapore (SEA), Tokyo (JP), Hong Kong (HK), Mumbai (IN), Frankfurt | ✅ Riot Vanguard |
| **Counter-Strike 2 (CS2)** | Tactical FPS | Singapore, Hong Kong, Tokyo, Seoul, Frankfurt (Valve SDR) | ✅ Valve Anti-Cheat (VAC) |
| **PUBG: BATTLEGROUNDS** | Battle Royale | Singapore, Tokyo, Seoul, Frankfurt (Azure & AWS) | ✅ BattlEye + Zakynthos |
| **Apex Legends** | Battle Royale | Singapore, Tokyo, Taiwan, Oregon (EA Multiplay) | ✅ Easy Anti-Cheat |
| **The Finals** | Arena FPS | Singapore, Tokyo, Frankfurt | ✅ Easy Anti-Cheat |
| **Call of Duty: Warzone / MW3** | FPS / BR | Singapore, Tokyo, US-West (Demonware) | ✅ RICOCHET Anti-Cheat |
| **Delta Force: Hawk Ops** | Tactical Shooter | Singapore, Hong Kong (Tencent Cloud / AWS) | ✅ ACE Anti-Cheat |
| **Overwatch 2** | Hero Shooter | Singapore, Taiwan, Korea, Japan | ✅ Blizzard Defense Matrix |
| **Rainbow Six Siege** | Tactical Shooter | Singapore (SEAU), Japan East (Ubisoft Azure) | ✅ BattlEye |
| **League of Legends (LoL)** | MOBA | Việt Nam (VNG), Singapore (Riot Direct), Taiwan | ✅ Riot Vanguard |
| **Dota 2** | MOBA | Singapore (SEA), Japan (Valve SDR) | ✅ Valve VAC |

*Người dùng cũng có thể dễ dàng thêm Game hoặc dải CIDR tùy chỉnh ngay trên giao diện Web Dashboard.*

---

## 🏗️ Kiến trúc Kỹ thuật (Technical Architecture)

```
[ Game: Valorant / CS2 / PUBG ]
       │ Gửi UDP/TCP tới IP Game Server (ví dụ: 13.250.0.0/15)
       ▼
[ Windows Routing Table ]
       │ Chỉ dải IP đích Game Server ──> [ WinTun Adapter: Lagvex ] (10.88.0.2)
       │ Mọi IP khác (Discord, Web) ──> [ Default Physical Gateway ] (Mạng thường)
       ▼
[ Lagvex Client Engine (Windows) ]
       │ Đọc raw IPv4 packets từ WinTun Ring Buffer
       │ Bọc 9-byte header siêu nhẹ (SessionID + MsgType)
       │ Gửi qua UDP socket tới Relay VPS (Route /32 ghim qua gateway vật lý)
       ▼
══════════════ [ Đường truyền cáp quang quốc tế / VPS ] ══════════════
       ▼
[ Lagvex Relay Server (Go trên Linux VPS) ]
       │ Bóc 9-byte header, xác thực HMAC-SHA256, Anti-spoofing
       │ Đẩy raw IPv4 packet vào Linux TUN (/dev/net/tun: lagvex0)
       ▼
[ Linux Kernel (VPS) ]
       │ iptables NAT MASQUERADE + TCP MSS Clamping
       ▼
[ Game Server Đích (Singapore / Tokyo / Hong Kong...) ]
```

### Các nguyên tắc an toàn mạng:
- **Pin Relay Route `/32`**: Luôn ghim địa chỉ IP của VPS Relay qua card mạng vật lý trước khi kích hoạt tunnel. Điều này triệt tiêu hoàn toàn nguy cơ **Routing Loop** (gói tin gửi tới relay bị hút ngược vào tunnel làm rớt mạng).
- **On-link Game Routes**: Các dải IP của game được định tuyến dưới dạng *on-link* (không nexthop), giúp Windows không mất thời gian chờ phân giải ARP/NDIS và không bị rớt gói.
- **`store=active` (RAM-only)**: Mọi thay đổi bảng định tuyến chỉ lưu trên RAM. Nếu máy tính tắt đột ngột, crash hoặc ngắt điện, toàn bộ route tự động biến mất 100%, không bao giờ gây lỗi mất mạng cho người dùng.

---

## 🚀 Hướng dẫn Cài đặt & Sử dụng

### 1. Dựng Relay Server trên VPS Linux (1 lệnh duy nhất)

Thuê một VPS Linux (Ubuntu / Debian / CentOS / Rocky / AlmaLinux) ở vị trí tối ưu (khuyến nghị **Singapore** hoặc **Tokyo/Hong Kong** từ các nhà cung cấp có đường truyền thẳng về Việt Nam như Vultr, Linode, AWS, DigitalOcean, Oracle Cloud, v.v.).

Chạy lệnh sau với quyền root:

```bash
curl -fsSL https://raw.githubusercontent.com/lagvex/lagvex/main/scripts/install-relay.sh | sudo bash
```

Script sẽ tự động:
1. Cấu hình `sysctl` tối ưu bộ đệm UDP và bật IP Forwarding (`net.ipv4.ip_forward = 1`).
2. Cấu hình tường lửa `iptables` / `ufw` NAT MASQUERADE và MSS clamping tự động lưu qua các lần reboot.
3. Tạo Pre-Shared Key (PSK) ngẫu nhiên 32 ký tự bảo mật.
4. Cài đặt và kích hoạt systemd daemon `lagvex-relay.service`.
5. In ra thông số IP, Port, PSK sẵn sàng để copy vào Client!

*(Bạn cũng có thể chạy qua Docker bằng `docker-compose up -d`)*

---

### 2. Chạy Client trên Windows

1. Tải bản build `lagvex-client.exe` và thư mục `web/`, `configs/` từ Releases.
2. Đảm bảo file `wintun.dll` (bản 64-bit) nằm cùng thư mục hoặc trong `bin/amd64/`.
3. Chạy `lagvex-client.exe` với quyền **Administrator** (bắt buộc để Windows cho phép tạo adapter mạng ảo WinTun và thêm route `store=active`):

```cmd
# Chạy với giao diện Web Dashboard (mặc định mở trình duyệt http://127.0.0.1:18888):
lagvex-client.exe

# Hoặc kết nối trực tiếp qua giao diện dòng lệnh (CLI):
lagvex-client.exe -connect -relay 123.45.67.89:51820 -psk your_secret_psk -game valorant -region asia-sg
```

---

## 🖥️ Giao diện Web Dashboard

Giao diện Dashboard hiện đại phong cách Cyberpunk Dark Neon:
- **Đo Ping thời gian thực (RTT Latency)** liên tục tới Relay Server.
- **Biểu đồ băng thông Up/Down** tính bằng KB/s hoặc MB/s.
- **Smart Process Watcher**: Tự động nhận diện khi tiến trình game chạy (ví dụ `VALORANT-Win64-Shipping.exe` bật lên) để nạp route, và tự gỡ route khi bạn tắt game.
- **Thư viện Game đa dạng**: Tìm kiếm và chọn cụm server của game chỉ bằng 1 cú click.
- **Ping Tester**: Kiểm tra tốc độ mạng tới VPS trước khi kết nối.

---

## 🛠️ Biên dịch từ mã nguồn (Build from Source)

Yêu cầu máy cài đặt **Go 1.22+**:

```bash
# Clone repository
git clone https://github.com/lagvex/lagvex.git
cd lagvex

# Chạy Unit Tests kiểm tra giao thức và HMAC
go test -v ./pkg/protocol

# Biên dịch Relay Server (cho Linux VPS)
make relay          # Tạo file bin/lagvex-relay

# Biên dịch Client (cho Windows)
make client         # Tạo file bin/lagvex-client.exe
```

---

## 📜 Giấy phép bản quyền (License)

Dự án được phát hành theo giấy phép mã nguồn mở **MIT License**. Bạn hoàn toàn tự do sử dụng, chỉnh sửa, tự host hoặc đóng góp phát triển.
