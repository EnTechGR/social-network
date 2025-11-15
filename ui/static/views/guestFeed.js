import { renderLayout } from "./components/layout.js";
import { renderPosts, mergePostsFromCategories } from "./components/posts.js";

const API_FEED = "http://localhost:8080/forum/api/feed";

export function renderGuestFeed(app) {
  // Render the shared app shell
  renderLayout(app, {
    isUser: false,
    mainContent: `
      <div class="feed-page fade-in">
        <div id="forumContainer" class="forum-container">
          <div class="loader">Loading feed...</div>
        </div>
      </div>
    `,
  });

  const forumContainer = document.getElementById("forumContainer");

  async function loadFeed() {
    try {
      const resp = await fetch(API_FEED, { credentials: "include" });
      if (!resp.ok) throw new Error("Failed to load feed");

      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []).sort(
        (a, b) => new Date(b.created_at) - new Date(a.created_at)
      );

      if (!posts.length) {
        forumContainer.innerHTML = `<p class="empty-feed">No posts available yet.</p>`;
        return;
      }

      renderPosts(forumContainer, posts, "/guest");
    } catch (err) {
      console.error("Error loading feed:", err);
      forumContainer.innerHTML = `
        <div class="error-message">
          <p>⚠️ Unable to load the feed. Please check your connection.</p>
          <button id="retryFeed" class="retry-btn">Retry</button>
        </div>
      `;
      document
        .getElementById("retryFeed")
        ?.addEventListener("click", () => loadFeed());
    }
  }

  loadFeed();
}
