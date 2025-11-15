import { renderPosts } from "../views/components/posts.js"; // reuse existing post cards

const API_BASE = "http://localhost:8080/forum/api";

export async function renderMyComments(main) {
  main.innerHTML = `
    <section class="my-comments-page fade-in">
      <header class="my-comments-header">
        <h1 class="my-comments-title">My Comments</h1>
        <p class="my-comments-description">
          Posts you’ve engaged with — your thoughts shared across BookTalk.
        </p>
      </header>

      <div id="commentedPostsContainer" class="forum-container">
        <div class="loader">Loading your commented posts...</div>
      </div>
    </section>
  `;

  const container = document.getElementById("commentedPostsContainer");

  async function loadCommentedPosts() {
    try {
      const resp = await fetch(`${API_BASE}/user/commented`, {
        credentials: "include",
      });

      if (!resp.ok) {
        const err = await resp.json().catch(() => ({}));
        throw new Error(err.message || "Failed to load commented posts");
      }

      const posts = await resp.json();

      if (!Array.isArray(posts) || posts.length === 0) {
        container.innerHTML = `<p class="empty-feed">You have not commented on any posts yet.</p>`;
        return;
      }

      renderPosts(container, posts, "/user");
    } catch (err) {
      console.error("Error loading commented posts:", err);
      container.innerHTML = `
        <div class="error-message">
          <p>⚠️ Unable to load your commented posts.</p>
        </div>
      `;
    }
  }

  loadCommentedPosts();
}
