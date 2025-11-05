import { navigateTo } from "../router.js";
import { renderPosts, mergePostsFromCategories } from "./components/posts.js";

const API_BASE = "http://localhost:8080/forum/api";

export async function renderCategoryPage(main, categoryId) {
  main.innerHTML = `
    <section class="category-page fade-in">
      <div class="category-header">
        <h2 class="category-title">Loading category...</h2>
        <p class="category-description"></p>
        <div class="category-toolbar">
          <select id="sortSelect" class="sort-select">
            <option value="newest">Newest</option>
            <option value="liked">Most Liked</option>
            <option value="commented">Most Commented</option>
          </select>
          <button id="createPostBtn" class="btn-accent">+ New Post</button>
        </div>
      </div>
      <div id="categoryPosts" class="forum-container">
        <div class="loader">Loading posts...</div>
      </div>
    </section>
  `;

  const titleEl = main.querySelector(".category-title");
  const descEl = main.querySelector(".category-description");
  const postsContainer = main.querySelector("#categoryPosts");
  const sortSelect = main.querySelector("#sortSelect");
  const createPostBtn = main.querySelector("#createPostBtn");

  try {
    // ✅ Fetch category info AND feed together
    const [categoryResp, feedResp] = await Promise.all([
      fetch(`${API_BASE}/category?id=${categoryId}`, {
        credentials: "include",
      }),
      fetch(`${API_BASE}/feed`, { credentials: "include" }),
    ]);

    if (!categoryResp.ok) {
      const err = await categoryResp.json().catch(() => ({}));
      throw new Error(err.message || "Failed to load category");
    }

    const category = await categoryResp.json();
    const feedData = await feedResp.json();

    titleEl.textContent = category.name || `Category ${categoryId}`;
    descEl.textContent = category.description || "";

    // Merge reactions/comments using feed data
    const allPosts = mergePostsFromCategories(feedData.categories || []);
    const posts =
      category.posts?.map((p) => {
        const enriched = allPosts.find((fp) => fp.id === p.id);
        return enriched ? enriched : p;
      }) || [];

    if (!posts.length) {
      postsContainer.innerHTML = `<p class="empty-feed">No posts in this category yet.</p>`;
      return;
    }

    renderPosts(postsContainer, posts, "/user");

    // Sorting support
    sortSelect.addEventListener("change", () => {
      const sorted = sortPosts([...posts], sortSelect.value);
      renderPosts(postsContainer, sorted, "/user");
    });

    createPostBtn.addEventListener("click", () => {
      navigateTo("/user/posts/create");
    });
  } catch (err) {
    console.error("Error loading category:", err);
    postsContainer.innerHTML = `<p class="error-message">⚠️ ${err.message}</p>`;
  }
}

function sortPosts(posts, criteria) {
  switch (criteria) {
    case "liked":
      return posts.sort(
        (a, b) =>
          (b.reactions?.filter((r) => r.reaction_type === 1).length || 0) -
          (a.reactions?.filter((r) => r.reaction_type === 1).length || 0)
      );
    case "commented":
      return posts.sort(
        (a, b) => (b.comment_count || 0) - (a.comment_count || 0)
      );
    default:
      return posts.sort(
        (a, b) => new Date(b.created_at) - new Date(a.created_at)
      );
  }
}
