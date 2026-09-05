const API_BASE = "/api/v1";

function formatApiError(payload, fallback) {
  if (!payload || typeof payload !== "object") return fallback;
  const detail = payload.error ?? payload.detail ?? payload.errors;
  if (Array.isArray(detail)) {
    const joined = detail.map(String).map((s) => s.trim()).filter(Boolean).join("; ");
    if (joined) return joined;
  } else if (detail != null && String(detail).trim()) {
    return String(detail).trim();
  }
  if (payload.message && String(payload.message).trim()) {
    return String(payload.message).trim();
  }
  return fallback;
}

async function api(method, path, body) {
  const opts = {
    method,
    headers: {},
  };
  if (body !== undefined) {
    opts.headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(API_BASE + path, opts);
  if (res.status === 204) {
    return { ok: true, status: 204, data: null };
  }
  let payload = null;
  const text = await res.text();
  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      payload = { message: text };
    }
  }
  if (!res.ok) {
    const msg = formatApiError(payload, res.statusText || "Request failed");
    const err = new Error(msg);
    err.status = res.status;
    err.payload = payload;
    throw err;
  }
  return {
    ok: true,
    status: res.status,
    data: payload && Object.prototype.hasOwnProperty.call(payload, "data")
      ? payload.data
      : payload,
  };
}

window.api = api;
