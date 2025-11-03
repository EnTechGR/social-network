const API_BASE = "http://localhost:8080/forum/api";

let currentUser = null;
let csrfToken = null;
let sessionChecked = false; // ✅ track if verifySession has run at least once

export async function verifySession() {
  try {
    const res = await fetch(`${API_BASE}/session/verify`, {
      credentials: "include",
    });

    if (!res.ok) {
      currentUser = null;
      csrfToken = null;
      sessionChecked = true;
      return null;
    }

    const data = await res.json();
    currentUser = data.user || null;
    csrfToken = data.csrf_token || null;
    sessionChecked = true;
    return currentUser;
  } catch (err) {
    console.error("Error verifying session:", err);
    clearSession();
    sessionChecked = true;
    return null;
  }
}

export function getCSRF() {
  return csrfToken;
}

export function getUser() {
  return currentUser;
}

export function isAuthenticated() {
  return !!currentUser;
}

export function clearSession() {
  currentUser = null;
  csrfToken = null;
  sessionChecked = false;
}

// ✅ Optional helper: returns a promise that resolves once session has been checked
export async function ensureSessionChecked() {
  if (sessionChecked) return;
  await verifySession();
}
