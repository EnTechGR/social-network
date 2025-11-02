import { renderWelcome } from "./views/welcome.js";
import { renderLogin } from "./views/login.js";
import { renderRegister } from "./views/register.js";
import { renderGuestFeed } from "./views/guestFeed.js";

export function router() {
  let path = window.location.pathname.replace(/\/+$/, "") || "/";
  const app = document.getElementById("app");

  switch (true) {
    case path === "/":
      renderWelcome(app);
      break;
    case path === "/login":
      renderLogin(app);
      break;
    case path === "/register":
      renderRegister(app);
      break;
    case path === "/guest/feed":
      renderGuestFeed(app);
      break;
    default:
      renderNotFound(app);
  }
}

// Navigate programmatically
export function navigateTo(path) {
  history.pushState({}, "", path);
  router();
}
