// /static/components/initLayout.js
import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";
import { clearSession } from "../../session.js";

export function initLayout() {
  const app = document.getElementById("app");

  // Render the base shell (only once)
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
          <div class="chat-placeholder">
            <p>💬 Live Chat (Coming Soon)</p>
          </div>
        </aside>
      </div>
    </div>
  `;

  // Sidebar categories
  const catDropdown = initCategoryDropdown({
    dropdownId: "category-list",
    basePath: "/user",
  });
  catDropdown.loadCategories();

  // Header actions
  document.getElementById("homeBtn").addEventListener("click", () => {
    navigateTo("/user/feed");
  });

  document.getElementById("logoutBtn").addEventListener("click", async () => {
    await fetch("http://localhost:8080/forum/api/session/logout", {
      method: "POST",
      credentials: "include",
    });
    clearSession();
    navigateTo("/login");
  });
}
