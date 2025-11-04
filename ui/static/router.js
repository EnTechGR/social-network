import { verifySession, isAuthenticated } from "./session.js";
import { renderLogin } from "./views/login.js";
import { renderRegister } from "./views/register.js";
import { renderWelcome } from "./views/welcome.js";
import { renderUserFeed } from "./views/userFeed.js";
// import { renderPost } from "./views/post.js";
// import { renderProfile } from "./views/profile.js";
// import { renderNotifications } from "./views/notifications.js";

let isNavigating = false;

export async function router() {
  const path = window.location.pathname.replace(/\/+$/, "") || "/";
  const app = document.getElementById("app");

  app.innerHTML = `<div class="loader">Loading...</div>`;

  await verifySession();
  const loggedIn = isAuthenticated();

  switch (true) {
    case path === "/":
      // If already logged in, send straight to the feed
      if (loggedIn) return navigateTo("/user/feed");
      renderWelcome(app);
      break;

    case path === "/login":
      if (loggedIn) return navigateTo("/user/feed");
      renderLogin(app);
      break;

    case path === "/register":
      if (loggedIn) return navigateTo("/user/feed");
      renderRegister(app);
      break;

    // Primary app route (requires auth)
    case path === "/user/feed":
      if (!loggedIn) return navigateTo("/login");
      renderUserFeed(app);
      break;

    // Protected user-only routes
    case path.startsWith("/user/post/"):
    case path === "/user/profile":
    case path === "/user/notifications":
      if (!loggedIn) return navigateTo("/login");
      // if (path.startsWith("/user/post/")) renderPost(app, path.split("/")[3]);
      // else if (path === "/user/profile") renderProfile(app);
      // else renderNotifications(app);
      app.innerHTML = `<h1>Coming soon</h1>`;
      break;

    default:
      app.innerHTML = `<h1>404 - Page Not Found</h1>`;
  }
}

export async function navigateTo(path) {
  if (isNavigating) return;
  isNavigating = true;
  history.pushState({}, "", path);
  await router();
  isNavigating = false;
}
