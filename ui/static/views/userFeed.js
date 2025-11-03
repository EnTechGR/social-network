import { renderLayout } from "./components/layout.js";
import { renderPosts, mergePostsFromCategories } from "./components/posts.js";
import { navigateTo } from "../router.js";

const API_FEED = "http://localhost:8080/forum/api/feed";

export function renderUserFeed(app) {
  // Render the shared app shell
  renderLayout(app, {
    isUser: true, // enables logout, notifications, and user menu
    mainContent: `
      <div class="feed-page">
        <div class="feed-toolbar">
          <h2>Feed</h2>
          <button id="createPostBtn" class="btn-accent">+ Create Post</button>
        </div>

        <div id="forumContainer" class="forum-container">
          <div class="loader">Loading feed...</div>
        </div>
      </div>
    `,
  });

  const forumContainer = document.getElementById("forumContainer");
  const createPostBtn = document.getElementById("createPostBtn");

  // Navigate to post creation page
  createPostBtn.addEventListener("click", () => {
    navigateTo("/user/posts/create");
  });

  // ✅ Fetch and render feed posts
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

      renderPosts(forumContainer, posts, "/user");
    } catch (err) {
      console.error("Error loading user feed:", err);
      forumContainer.innerHTML = `<p class="error-message">⚠️ Unable to load feed. Please try again later.</p>`;
    }
  }

  loadFeed();
}
