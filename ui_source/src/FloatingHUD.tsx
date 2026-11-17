import { useCallback, useEffect, useRef, useState } from "react";
import type { PointerEvent as ReactPointerEvent } from "react";
import "./FloatingHUD.css";

const HUD_VISIBLE_KEY = "lagvex_hud_visible";
const HUD_POS_KEY = "lagvex_hud_pos";
const POLL_INTERVAL_MS = 2000;

type StatusSnapshot = {
  active: boolean;
  state: string;
  pingMs: number | null;
  lossPct: number | null;
  node: string;
};

type Position = { left: number; top: number };

const EMPTY_STATUS: StatusSnapshot = {
  active: false,
  state: "unavailable",
  pingMs: null,
  lossPct: null,
  node: "Unavailable",
};

function readVisible(): boolean {
  try {
    return localStorage.getItem(HUD_VISIBLE_KEY) === "true";
  } catch {
    return false;
  }
}

function readPosition(): Position | null {
  try {
    const raw = localStorage.getItem(HUD_POS_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<Position>;
    if (
      typeof parsed.left !== "number" ||
      typeof parsed.top !== "number" ||
      !Number.isFinite(parsed.left) ||
      !Number.isFinite(parsed.top)
    ) {
      return null;
    }
    return { left: parsed.left, top: parsed.top };
  } catch {
    return null;
  }
}

function persistPosition(position: Position) {
  try {
    localStorage.setItem(HUD_POS_KEY, JSON.stringify(position));
  } catch {
    // Position persistence is optional; dragging still works for this session.
  }
}

/**
 * Optional floating latency pill. It reads only the real local engine status
 * endpoint and renders an explicit unavailable state when a value has not been
 * measured yet. It never invents latency, loss, or node names.
 */
export default function FloatingHUD() {
  const [visible, setVisible] = useState(readVisible);
  const [minimized, setMinimized] = useState(false);
  const [position, setPosition] = useState<Position | null>(readPosition);
  const [status, setStatus] = useState<StatusSnapshot>(EMPTY_STATUS);

  const rootRef = useRef<HTMLDivElement | null>(null);
  const dragRef = useRef<{
    pointerId: number;
    startX: number;
    startY: number;
    originLeft: number;
    originTop: number;
  } | null>(null);

  const clampPosition = useCallback((next: Position): Position => {
    const root = rootRef.current;
    if (!root) return next;
    const maxLeft = Math.max(8, window.innerWidth - root.offsetWidth - 8);
    const maxTop = Math.max(8, window.innerHeight - root.offsetHeight - 8);
    return {
      left: Math.min(Math.max(8, next.left), maxLeft),
      top: Math.min(Math.max(8, next.top), maxTop),
    };
  }, []);

  useEffect(() => {
    const onResize = () => {
      setPosition((current) => (current ? clampPosition(current) : current));
    };
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, [clampPosition]);

  useEffect(() => {
    if (!visible) return;

    let cancelled = false;
    let timer: number | undefined;

    const poll = async () => {
      try {
        const res = await fetch("/api/status", { cache: "no-store" });
        if (!res.ok) throw new Error(`status ${res.status}`);
        const data = (await res.json()) as {
          active?: boolean;
          state?: string;
          pingMs?: number;
          packetLossPct?: number;
          activeRelayName?: string;
          activeRelayId?: string;
        };
        if (cancelled) return;

        const connected = data.active === true || data.state === "connected";
        const pingMs =
          connected && typeof data.pingMs === "number" && data.pingMs > 0
            ? data.pingMs
            : null;
        const lossPct =
          connected && typeof data.packetLossPct === "number"
            ? data.packetLossPct
            : null;

        let node = "Disconnected";
        if (connected) {
          node = data.activeRelayName || data.activeRelayId || "Relay";
        } else if (data.state === "error") {
          node = "Relay error";
        }

        setStatus({
          active: connected,
          state: typeof data.state === "string" ? data.state : "unknown",
          pingMs,
          lossPct,
          node,
        });
      } catch {
        if (!cancelled) setStatus(EMPTY_STATUS);
      } finally {
        if (!cancelled) timer = window.setTimeout(poll, POLL_INTERVAL_MS);
      }
    };

    poll();
    return () => {
      cancelled = true;
      if (timer !== undefined) window.clearTimeout(timer);
    };
  }, [visible]);

  const setHudVisible = (next: boolean) => {
    setVisible(next);
    try {
      localStorage.setItem(HUD_VISIBLE_KEY, next ? "true" : "false");
    } catch {
      // The HUD still toggles for this session when storage is blocked.
    }
  };

  const handlePointerDown = (event: ReactPointerEvent<HTMLDivElement>) => {
    if ((event.target as HTMLElement).closest("button")) return;
    const root = rootRef.current;
    if (!root) return;
    const rect = root.getBoundingClientRect();
    dragRef.current = {
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY,
      originLeft: rect.left,
      originTop: rect.top,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
    event.preventDefault();
  };

  const handlePointerMove = (event: ReactPointerEvent<HTMLDivElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    setPosition(
      clampPosition({
        left: drag.originLeft + event.clientX - drag.startX,
        top: drag.originTop + event.clientY - drag.startY,
      }),
    );
  };

  const endDrag = (event: ReactPointerEvent<HTMLDivElement>) => {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    dragRef.current = null;
    const rect = rootRef.current?.getBoundingClientRect();
    if (rect) persistPosition({ left: rect.left, top: rect.top });
  };

  const pingClass =
    status.pingMs === null
      ? ""
      : status.pingMs < 35
        ? "is-good"
        : status.pingMs < 75
          ? "is-mid"
          : "is-high";

  const lossClass =
    status.lossPct === null
      ? ""
      : status.lossPct === 0
        ? "is-good"
        : status.lossPct < 2
          ? "is-mid"
          : "is-high";

  return (
    <>
      {visible && (
        <div
          ref={rootRef}
          className="lvx-hud-root"
          style={position ? { top: position.top, left: position.left, right: "auto" } : undefined}
        >
          <div
            className="lvx-hud-pill"
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={endDrag}
            onPointerCancel={endDrag}
          >
            <span className="lvx-hud-drag" title="Drag to reposition" aria-hidden="true">
              ⠿
            </span>
            <span
              className={`lvx-hud-dot${
                status.active ? " is-online" : status.state === "error" ? " is-error" : ""
              }`}
              title={status.active ? "Tunnel connected" : `Tunnel ${status.state}`}
            />
            <div className="lvx-hud-metric">
              <span className="lvx-hud-label">PING</span>
              <span
                className={`lvx-hud-value ${pingClass}`}
                title="Round-trip time measured by the relay keepalive"
              >
                {status.pingMs === null ? "—" : `${status.pingMs.toFixed(0)} ms`}
              </span>
            </div>
            {!minimized && (
              <>
                <div className="lvx-hud-divider" />
                <div className="lvx-hud-metric">
                  <span className="lvx-hud-label">LOSS</span>
                  <span
                    className={`lvx-hud-value ${lossClass}`}
                    title="Estimated from missed relay keepalives; not a packet capture"
                  >
                    {status.lossPct === null ? "—" : `${status.lossPct.toFixed(1)}%`}
                  </span>
                </div>
                <div className="lvx-hud-divider" />
                <div className="lvx-hud-metric">
                  <span className="lvx-hud-label">NODE</span>
                  <span className="lvx-hud-value lvx-hud-node" title={status.node}>
                    {status.node}
                  </span>
                </div>
              </>
            )}
            <div className="lvx-hud-controls">
              <button
                type="button"
                className="lvx-hud-button"
                onClick={() => setMinimized((value) => !value)}
                title={minimized ? "Expand HUD" : "Minimize HUD"}
                aria-label={minimized ? "Expand HUD" : "Minimize HUD"}
              >
                {minimized ? "+" : "−"}
              </button>
              <button
                type="button"
                className="lvx-hud-button"
                onClick={() => setHudVisible(false)}
                title="Close HUD"
                aria-label="Close HUD"
              >
                ×
              </button>
            </div>
          </div>
        </div>
      )}

      <button
        type="button"
        className={`lvx-hud-launcher${visible ? " is-open" : ""}`}
        onClick={() => setHudVisible(!visible)}
        title={visible ? "Hide the floating latency HUD" : "Show the floating latency HUD"}
        aria-pressed={visible}
      >
        <svg aria-hidden="true" className="lvx-hud-launcher-bolt" width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
        </svg>
        <span>HUD</span>
      </button>
    </>
  );
}
