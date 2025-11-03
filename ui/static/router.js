import { verifySession, isAuthenticated } from "./session.js";
import { renderLogin } from "./views/login.js";
import { renderRegister } from "./views/register.js";
import { renderWelcome } from "./views/welcome.js";
import { renderGuestFeed } from "./views/guestFeed.js";
import { renderUserFeed } from "./views/userFeed.js";
// import { renderPost } from "./views/post.js";
// import { renderProfile } from "./views/profile.js";
// import { renderNotifications } from "./views/notifications.js";

export async function router() {
  const path = window.location.pathname;
  const app = document.getElementById("app");

  // Always check session before deciding where to go
  await verifySession();
  const loggedIn = isAuthenticated();

  switch (true) {
    // Root → Welcome
    case path === "/":
      renderWelcome(app);
      break;

    // Auth pages (guest only)
    case path === "/login":
      if (loggedIn) return navigateTo("/user/feed");
      renderLogin(app);
      break;

    case path === "/register":
      if (loggedIn) return navigateTo("/user/feed");
      renderRegister(app);
      break;

    // Guest feed (redirect user to user feed)
    case path === "/guest/feed":
      if (loggedIn) return navigateTo("/user/feed");
      renderGuestFeed(app);
      break;

    // User feed (redirect guest to login)
    case path === "/user/feed":
      if (!loggedIn) return navigateTo("/login");
      renderUserFeed(app);
      break;

    // Protected user-only routes
    case path.startsWith("/user/post/"):
    case path === "/user/profile":
    case path === "/user/notifications":
      if (!loggedIn) return navigateTo("/login");
      if (path.startsWith("/user/post/")) {
        renderPost(app, path.split("/")[3]);
      } else if (path === "/user/profile") {
        renderProfile(app);
      } else {
        renderNotifications(app);
      }
      break;

    default:
      app.innerHTML = `<h1>404 - Page Not Found</h1>`;
  }
}

// Programmatic navigation
export function navigateTo(path) {
  history.pushState({}, "", path);
  router();
}
