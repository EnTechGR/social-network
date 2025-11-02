import { initCategoryDropdown } from "./components/categories.js";
import { renderPosts, mergePostsFromCategories } from "./components/posts.js";

const API_FEED = "http://localhost:8080/forum/api/feed";

export function renderGuestFeed(app) {
  app.innerHTML = `
    <div class="feed-page">
      <button class="category-dropdown-toggle">Categories <span class="dropdown-arrow">▼</span></button>
      <ul id="category-tabs" class="category-dropdown"></ul>
      <div id="forumContainer" class="forum-container"></div>
    </div>
  `;

  const forumContainer = document.getElementById("forumContainer");

  // Initialize category dropdown
  const categoryDropdown = initCategoryDropdown({
    toggleSelector: ".category-dropdown-toggle",
    dropdownId: "category-tabs",
    basePath: "/guest",
  });

  categoryDropdown.loadCategories();

  async function loadFeed() {
    try {
      const resp = await fetch(API_FEED, { credentials: "include" });
      if (!resp.ok) throw new Error("Failed to load feed");
      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []).sort(
        (a, b) => new Date(b.created_at) - new Date(a.created_at)
      );
      renderPosts(forumContainer, posts, "/guest");
    } catch (err) {
      console.error("Error loading feed:", err);
      forumContainer.textContent = "Failed to load feed";
    }
  }

  loadFeed();
}
