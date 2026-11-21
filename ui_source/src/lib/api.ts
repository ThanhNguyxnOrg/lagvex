// Shared guarded fetch wrapper for the local Lagvex dashboard API.
// Every state-changing request carries X-Lagvex-Request: 1 so cross-site web
// pages cannot drive the loopback server without a CORS preflight (which the
// server intentionally never grants).

const nativeFetch = window.fetch.bind(window);

export function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers);
  headers.set("X-Lagvex-Request", "1");
  return nativeFetch(input, { ...init, headers });
}
