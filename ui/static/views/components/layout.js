import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";

// Renders the shared forum layout shell
export function renderLayout(app, { isUser = false, mainContent = "" }) {
  app.innerHTML = `
    <div class="layout">
      <!-- Header -->
      <header class="app-header">
        <h1 class="logo" id="homeBtn">Forum</h1>
        <div class="header-actions">
          ${isUser ? '<button id="notificationsBtn">🔔</button>' : ""}
          <button id="logoutBtn">Logout</button>
        </div>
      </header>

      <div class="app-body">
        <!-- Sidebar -->
        <aside class="sidebar">
          <h2>Categories</h2>
          <ul id="category-list" class="category-list"></ul>
          ${
            isUser
              ? `
            <div class="user-menu">
              <h3>Your Space</h3>
              <ul>
                <li><a href="/user/my-activity/my-posts">My Posts</a></li>
                <li><a href="/user/my-activity/my-reactions">My Reactions</a></li>
                <li><a href="/user/my-activity/my-comments">My Comments</a></li>
              </ul>
            </div>
          `
              : ""
          }
        </aside>

        <!-- Main Content -->
        <main class="main-content" id="mainContent">
          ${mainContent}
        </main>

        <!-- Right Panel (chat placeholder) -->
        <aside class="right-panel">
          <div class="chat-placeholder">
            <p>💬 Live Chat (Coming Soon)</p>
          </div>
        </aside>
      </div>
    </div>
  `;

  // Hook up interactions
  document.getElementById("homeBtn").addEventListener("click", () => {
    navigateTo(isUser ? "/user/feed" : "/guest/feed");
  });

  const logoutBtn = document.getElementById("logoutBtn");
  logoutBtn.addEventListener("click", async () => {
    try {
      await fetch("http://localhost:8080/forum/api/session/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch (err) {
      console.error("Logout failed:", err);
    }
    navigateTo("/login");
  });

  // Load categories into sidebar
  const catDropdown = initCategoryDropdown({
    toggleSelector: null, // no dropdown here
    dropdownId: "category-list",
    basePath: isUser ? "/user" : "/guest",
  });
  catDropdown.loadCategories();
}
