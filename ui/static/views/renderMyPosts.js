import { renderPosts } from "./components/posts.js";

const API_BASE = "http://localhost:8080/forum/api";

export function renderMyPosts(main) {
  main.innerHTML = `
    <section class="my-posts-page fade-in">
      <header class="my-posts-header">
        <h1 class="my-posts-title">My Posts</h1>
        <p class="my-posts-description">
          A collection of your writings and reflections shared with BookTalk.
        </p>
      </header>

      <div id="myPostsContainer" class="forum-container">
        <div class="loader">Loading your posts...</div>
      </div>
    </section>
  `;

  const container = document.getElementById("myPostsContainer");

  loadMyPosts();

  async function loadMyPosts() {
    try {
      const resp = await fetch(`${API_BASE}/user/posts`, {
        credentials: "include",
      });
      if (!resp.ok) throw new Error("Failed to load posts");

      const posts = await resp.json();
      if (!Array.isArray(posts) || posts.length === 0) {
        container.innerHTML = `<p class="empty-feed">You haven’t written any posts yet.</p>`;
        return;
      }

      // 🔥 Use shared renderer:
      // - Clicking a card → /user/my-activity/my-posts/edit/post/:id
      // - Category links → /user/category/:id
      renderPosts(container, posts, {
        postBasePath: "/user/my-activity/my-posts/edit",
        categoryBasePath: "/user",
      });
    } catch (err) {
      console.error("Error loading posts:", err);
      container.innerHTML = `<p class="error-message">Unable to load your posts.</p>`;
    }
  }
}
