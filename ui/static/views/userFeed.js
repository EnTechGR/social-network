import { renderPosts, mergePostsFromCategories } from "./components/posts.js";
import { navigateTo } from "../router.js";

const API_FEED = "http://localhost:8080/forum/api/feed";

export function renderUserFeed(main) {
  // Render only the inner content — not the full layout
  main.innerHTML = `
    <div class="feed-page fade-in">
      <div class="feed-toolbar">
        <h2>Latest Posts</h2>
      </div>

      <div id="forumContainer" class="forum-container">
        <div class="loader">Loading your personalized feed...</div>
      </div>
    </div>
  `;

  const forumContainer = document.getElementById("forumContainer");

  /* ========== LOAD FEED ========== */
  async function loadFeed() {
    try {
      const resp = await fetch(API_FEED, { credentials: "include" });
      if (!resp.ok) throw new Error("Failed to load feed");

      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []).sort(
        (a, b) => new Date(b.created_at) - new Date(a.created_at)
      );

      if (!posts.length) {
        forumContainer.innerHTML = `
          <p class="empty-feed">
            No posts yet. Be the first to share something!
          </p>
        `;
        return;
      }

      renderPosts(forumContainer, posts, "/user");
    } catch (err) {
      console.error("Error loading user feed:", err);
      forumContainer.innerHTML = `
        <div class="error-message">
          <p>⚠️ Unable to load your feed. Please check your connection.</p>
          <button id="retryFeed" class="retry-btn">Retry</button>
        </div>
      `;
      document.getElementById("retryFeed")?.addEventListener("click", loadFeed);
    }
  }

  loadFeed();
}
