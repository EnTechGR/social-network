import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";
import { clearSession } from "../../session.js";

export function renderLayout(app, { isUser = true, mainContent = "" }) {
  app.innerHTML = `
    <div class="layout fade-in">
      <header class="app-header">
        <h1 class="logo" id="homeBtn">📚 BookTalk</h1>
        <div class="header-actions">
          <button id="notificationsBtn" title="Notifications">🔔</button>
          <button id="logoutBtn" title="Logout">Logout</button>
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

        <main class="main-content" id="mainContent">
          ${mainContent}
        </main>

        <aside class="right-panel">
          <div class="chat-placeholder">
            <p>💬 Live Chat (Coming Soon)</p>
          </div>
        </aside>
      </div>
    </div>
  `;

  document.getElementById("homeBtn")?.addEventListener("click", () => {
    navigateTo("/user/feed");
  });

  document.getElementById("logoutBtn")?.addEventListener("click", async () => {
    try {
      await fetch("http://localhost:8080/forum/api/session/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch (err) {
      console.error("Logout failed:", err);
    } finally {
      clearSession();
      navigateTo("/");
    }
  });

  // Categories: basePath now always /user
  const catDropdown = initCategoryDropdown({
    toggleSelector: null,
    dropdownId: "category-list",
    basePath: "/user",
  });
  catDropdown.loadCategories();
}
