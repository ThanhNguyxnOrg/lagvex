#!/usr/bin/env bash
set -euo pipefail

export root="$1"
scenario="$2"
export archive=lagvex-relay-v1.2.3-linux-amd64.tar.gz
mkdir -p "$root"/{commands,release,bin,etc/sysctl.d,etc/systemd/system}
for cmd in awk bash cat chmod cp grep gzip mkdir mktemp mv rm sha256sum tar; do
    ln -s "$(command -v "$cmd")" "$root/commands/$cmd"
done

cat > "$root/commands/ip" <<'EOF'
#!/bin/bash
echo 'default via 192.0.2.1 dev eth0'
EOF
cat > "$root/commands/uname" <<'EOF'
#!/bin/bash
echo x86_64
EOF
cat > "$root/commands/curl" <<'EOF'
#!/bin/bash
set -eu
while (($#)); do
    case "$1" in
        -o) dest="$2"; shift ;;
        https://*) url="$1" ;;
    esac
    shift
done
case "$url" in
    https://api.ipify.org) echo '192.0.2.2' ;;
    https://github.com/ThanhNguyxnOrg/lagvex/releases/latest/download/checksums.txt)
        cp "$root/release/checksums.txt" "$dest" ;;
    "https://github.com/ThanhNguyxnOrg/lagvex/releases/download/v1.2.3/$archive")
        cp "$root/release/$archive" "$dest" ;;
    *) exit 22 ;;
esac
EOF
for cmd in modprobe sysctl iptables systemctl; do
    cat > "$root/commands/$cmd" <<'EOF'
#!/bin/bash
echo "${0##*/} $*" >> "$root/events"
EOF
done
chmod +x "$root"/commands/{curl,ip,uname,modprobe,sysctl,iptables,systemctl}

cat > "$root/release/lagvex-relay" <<'EOF'
#!/bin/bash
[[ "$1" == -version ]] || exit 1
echo version >> "$root/events"
EOF
(
    cd "$root/release"
    tar -czf "$archive" ./lagvex-relay
    sha256sum "$archive" > checksums.txt
    case "$scenario" in
        wrong-checksum) printf '%064d  %s\n' 0 "$archive" > checksums.txt ;;
        missing-archive) rm "$archive" ;;
        missing-checksum) : > checksums.txt ;;
    esac
)

# Exercise the documented stdin invocation without root, network, Go or host paths.
sed -e 's/\$EUID/0/g' -e "s|/etc/|$root/etc/|g" -e "s|/usr/local/bin|$root/bin|g" \
    -e 's|/dev/net/tun|/dev/null|g' install-relay.sh > "$root/installer"
if (cd "$root" && PATH="$root/commands" LAGVEX_PSK=test bash < installer); then
    [[ "$scenario" == valid ]]
    [[ -x "$root/bin/lagvex-relay" && "$(head -1 "$root/events")" == version ]]
    grep -Fx 'systemctl enable --now lagvex-relay.service' "$root/events"
else
    [[ "$scenario" != valid ]]
    [[ ! -e "$root/events" && -z "$(ls -A "$root/bin")" ]]
    [[ -z "$(find "$root/etc" -type f -print)" ]]
fi
