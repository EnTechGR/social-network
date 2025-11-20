// /static/components/initLayout.js
import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";
import { clearSession } from "../../session.js";
import {
  initLiveChatSidebar,
  disconnectLiveChatSidebar,
} from "../components/liveChatSidebar.js";

export function initLayout() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="layout">
      <header class="app-header">
        <h1 class="logo" id="homeBtn">BookTalk</h1>
        <div class="header-actions">
          <button id="notificationsBtn">🔔</button>
          <button id="logoutBtn">Logout</button>
        </div>
      </header>

      <div class="app-body">
        <aside class="sidebar">
          <h2>Categories</h2>
          <ul id="category-list" class="category-list"></ul>

          <div class="user-menu">
            <h3>Your Space</h3>
            <ul>
              <li><a href="/user/my-activity/my-posts">My Posts</a></li>
              <li><a href="/user/my-activity/my-reactions">My Reactions</a></li>
              <li><a href="/user/my-activity/my-comments">My Comments</a></li>
            </ul>
          </div>
        </aside>

        <main id="mainContent" class="main-content">
          <div class="loader">Loading...</div>
        </main>

        <aside class="right-panel">
          <div id="chatSidebar" class="chat-sidebar">
            <h3>Live Chat</h3>
            <div id="currentUserInfo" class="current-user-info">
              Loading...
            </div>

            <h4>Users</h4>
            <div id="usersList" class="chat-users-list"></div>
          </div>
        </aside>
      </div>
    </div>
  `;

  const catDropdown = initCategoryDropdown({
    dropdownId: "category-list",
    basePath: "/user",
  });
  catDropdown.loadCategories();

  document.getElementById("homeBtn").addEventListener("click", () => {
    navigateTo("/user/feed");
  });

  document.getElementById("notificationsBtn").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/user/notifications");
  });

  document.getElementById("logoutBtn").addEventListener("click", async () => {
    disconnectLiveChatSidebar();
    await fetch("http://localhost:8080/forum/api/session/logout", {
      method: "POST",
      credentials: "include",
    });
    clearSession();
    navigateTo("/login");
  });

  initLiveChatSidebar();
}
