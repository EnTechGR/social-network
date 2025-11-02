const API_BASE = "http://localhost:8080/forum/api";

let currentUser = null;
let csrfToken = null;

export async function verifySession() {
  const res = await fetch(`${API_BASE}/session/verify`, {
    credentials: "include",
  });

  if (res.ok) {
    const data = await res.json();
    currentUser = data.user;
    csrfToken = data.csrf_token;
    return currentUser;
  } else {
    currentUser = null;
    csrfToken = null;
    return null;
  }
}

export function getCSRF() {
  return csrfToken;
}

export function getUser() {
  return currentUser;
}
