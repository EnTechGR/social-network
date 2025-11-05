import { verifySession, isAuthenticated } from "./session.js";
import { initLayout } from "./views/components/layout.js";
import { renderLogin } from "./views/login.js";
import { renderRegister } from "./views/register.js";
import { renderWelcome } from "./views/welcome.js";
import { renderUserFeed } from "./views/userFeed.js";
import { renderCreatePost } from "./views/createPost.js";
import { renderPost } from "./views/renderPost.js";
import { renderCategoryPage } from "./views/renderCategoryPage.js";
import { renderMyPosts } from "./views/renderMyPosts.js";
import { renderMyReactions } from "./views/renderMyReactions.js";
import { renderMyComments } from "./views/renderMyComments.js";
// import { renderProfile } from "./views/profile.js";
// import { renderNotifications } from "./views/notifications.js";

let layoutInitialized = false;

export async function router() {
  const path = window.location.pathname.replace(/\/+$/, "") || "/";
  const app = document.getElementById("app");
  const main = document.getElementById("mainContent");

  await verifySession();
  const loggedIn = isAuthenticated();

  // Initialize layout only once (for logged-in users)
  if (loggedIn && !layoutInitialized) {
    initLayout();
    layoutInitialized = true;
  }

  // Determine where to render
  const target = loggedIn ? document.getElementById("mainContent") : app;
  if (target) target.innerHTML = `<div class="loader">Loading...</div>`;

  switch (true) {
    case path === "/":
      if (loggedIn) return navigateTo("/user/feed");
      renderWelcome(target);
      break;

    case path === "/login":
      if (loggedIn) return navigateTo("/user/feed");
      renderLogin(target);
      break;

    case path === "/register":
      if (loggedIn) return navigateTo("/user/feed");
      renderRegister(target);
      break;

    case path === "/user/feed":
      if (!loggedIn) return navigateTo("/login");
      renderUserFeed(target);
      break;

    case path === "/user/posts/create":
      if (!loggedIn) return navigateTo("/login");
      renderCreatePost(target);
      break;

    case path.startsWith("/user/post/"):
      if (!loggedIn) return navigateTo("/login");
      const postId = path.split("/").pop();
      renderPost(target, postId);
      break;

    case path.startsWith("/user/category/"):
      if (!loggedIn) return navigateTo("/login");
      const categoryId = path.split("/").pop();
      renderCategoryPage(target, categoryId);
      break;

    case path === "/user/my-activity/my-posts":
      if (!loggedIn) return navigateTo("/login");
      renderMyPosts(target);
      break;

    case path === "/user/my-activity/my-reactions":
      if (!loggedIn) return navigateTo("/login");
      renderMyReactions(target);
      break;

    case path === "/user/my-activity/my-comments":
      if (!loggedIn) return navigateTo("/login");
      renderMyComments(target);
      break;

    // // Protected user-only routes
    // case path.startsWith("/user/post/"):
    // case path === "/user/profile":
    // case path === "/user/notifications":
    //   if (!loggedIn) return navigateTo("/login");
    //   // if (path.startsWith("/user/post/")) renderPost(app, path.split("/")[3]);
    //   // else if (path === "/user/profile") renderProfile(app);
    //   // else renderNotifications(app);
    //   app.innerHTML = `<h1>Coming soon</h1>`;
    //   break;

    default:
      target.innerHTML = `<h1>404 - Page Not Found</h1>`;
  }
}

export async function navigateTo(path) {
  history.pushState({}, "", path);
  await router();
}
