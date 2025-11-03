import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";
import { clearSession } from "../../session.js";

// Renders the shared BookTalk forum layout
export function renderLayout(app, { isUser = false, mainContent = "" }) {
  app.innerHTML = `
    <div class="layout fade-in">
      <!-- Header -->
      <header class="app-header">
        <h1 class="logo" id="homeBtn">📚 BookTalk</h1>
        <div class="header-actions">
          ${
            isUser
              ? `<button id="notificationsBtn" title="Notifications">🔔</button>`
              : ""
          }
          <button id="logoutBtn" title="Logout">Logout</button>
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

        <!-- Right Panel -->
        <aside class="right-panel">
          <div class="chat-placeholder">
            <p>💬 Live Chat (Coming Soon)</p>
          </div>
        </aside>
      </div>
    </div>
  `;

  /* ---------- HEADER INTERACTIONS ---------- */
  const homeBtn = document.getElementById("homeBtn");
  if (homeBtn)
    homeBtn.addEventListener("click", () =>
      navigateTo(isUser ? "/user/feed" : "/guest/feed")
    );

  const logoutBtn = document.getElementById("logoutBtn");
  if (logoutBtn)
    logoutBtn.addEventListener("click", async () => {
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

  /* ---------- SIDEBAR CATEGORIES ---------- */
  const catDropdown = initCategoryDropdown({
    toggleSelector: null, // no dropdown on sidebar
    dropdownId: "category-list",
    basePath: isUser ? "/user" : "/guest",
  });
  catDropdown.loadCategories();
}
