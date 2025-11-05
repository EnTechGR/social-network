import { navigateTo } from "../router.js";

const API_BASE = "http://localhost:8080/forum/api";

export function renderMyPosts(main) {
  main.innerHTML = `
    <section class="my-posts-page fade-in">
      <div class="page-header">
        <h2>Your Posts</h2>
        <button id="newPostBtn" class="btn-accent">+ New Post</button>
      </div>
      <div id="myPostsContainer" class="forum-container">
        <div class="loader">Loading your posts...</div>
      </div>
    </section>
  `;

  const container = document.getElementById("myPostsContainer");
  const newPostBtn = document.getElementById("newPostBtn");

  newPostBtn.addEventListener("click", () => navigateTo("/user/create-post"));

  loadMyPosts();

  async function loadMyPosts() {
    try {
      const resp = await fetch(`${API_BASE}/user/posts`, {
        credentials: "include",
      });
      if (!resp.ok) throw new Error("Failed to load posts");

      const posts = await resp.json();
      if (!posts.length) {
        container.innerHTML = `<p class="empty-feed">You haven’t written any posts yet.</p>`;
        return;
      }

      renderPosts(posts);
    } catch (err) {
      console.error("Error loading posts:", err);
      container.innerHTML = `<p class="error-message">Unable to load your posts.</p>`;
    }
  }

  function renderPosts(posts) {
    container.innerHTML = "";
    posts.forEach((post) => {
      const card = document.createElement("div");
      card.className = "post-card clickable-post";

      card.addEventListener("click", () => {
        navigateTo(`/user/my-activity/my-posts/edit/post/${post.id}`);
      });

      // Optional thumbnail
      if (post.thumbnail_url) {
        const img = document.createElement("img");
        img.src = post.thumbnail_url;
        img.alt = "Post thumbnail";
        img.className = "post-thumb";
        card.appendChild(img);
      }

      const title = document.createElement("h3");
      title.textContent =
        post.title ||
        (post.content ? post.content.slice(0, 30) + "..." : "Untitled Post");
      card.appendChild(title);

      const meta = document.createElement("p");
      const isEdited = post.updated_at && post.updated_at !== post.created_at;
      meta.textContent = `By You • ${new Date(
        post.updated_at || post.created_at
      ).toLocaleString()}${isEdited ? " (Edited)" : ""}`;
      card.appendChild(meta);

      const content = document.createElement("p");
      content.textContent =
        post.content?.length > 150
          ? post.content.slice(0, 150) + "..."
          : post.content || "No content.";
      card.appendChild(content);

      // Reaction + comments footer
      const likes = (post.reactions || []).filter(
        (r) => r.reaction_type === 1
      ).length;
      const dislikes = (post.reactions || []).filter(
        (r) => r.reaction_type === 2
      ).length;
      const commentCount =
        post.comment_count || (post.comments ? post.comments.length : 0);

      const footer = document.createElement("div");
      footer.className = "post-footer";
      footer.innerHTML = `👍 ${likes} | 👎 ${dislikes} | 💬 ${commentCount}`;
      card.appendChild(footer);

      container.appendChild(card);
    });
  }
}
