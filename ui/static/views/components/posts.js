import { navigateTo } from "../../router.js";

export function renderPosts(container, posts, basePath = "/user") {
  container.innerHTML = "";

  if (!posts || posts.length === 0) {
    container.textContent = "No posts available.";
    return;
  }

  posts.forEach((post) => {
    const wrapper = document.createElement("div");
    wrapper.className = "post-card clickable-post";
    wrapper.style.cursor = "pointer";

    // Navigate to post page
    wrapper.addEventListener("click", () => {
      navigateTo(`${basePath}/post/${post.id}`);
    });

    // Thumbnail (optional)
    if (post.thumbnail_url) {
      const img = document.createElement("img");
      img.src = post.thumbnail_url;
      img.alt = "Post thumbnail";
      img.className = "post-thumb";
      wrapper.appendChild(img);
    }

    // Title
    const header = document.createElement("h3");
    header.textContent = post.title || "Untitled Post";
    wrapper.appendChild(header);

    // Author and timestamp
    const author = document.createElement("p");
    const edited =
      post.updated_at && post.updated_at !== post.created_at ? " (Edited)" : "";
    const time = new Date(post.updated_at || post.created_at).toLocaleString();

    author.textContent = `By ${
      post.username || "Anonymous"
    } | ${time}${edited}`;
    wrapper.appendChild(author);

    // Content snippet
    const content = document.createElement("p");
    content.textContent = post.content || "";
    wrapper.appendChild(content);

    // Reaction counts
    const likes =
      post.reactions?.filter((r) => r.reaction_type === 1).length || 0;
    const dislikes =
      post.reactions?.filter((r) => r.reaction_type === 2).length || 0;
    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);

    const footer = document.createElement("div");
    footer.className = "post-footer";
    footer.textContent = `▲ ${likes} | ▼ ${dislikes} | 💬 ${commentCount}`;
    wrapper.appendChild(footer);

    // Categories
    if (post.categories && post.categories.length > 0) {
      const catContainer = document.createElement("div");
      catContainer.className = "post-categories";

      post.categories.forEach((cat, idx) => {
        const catLink = document.createElement("a");
        catLink.href = `${basePath}/category/${cat.id}`;
        catLink.textContent = cat.name;
        catLink.className = "post-category-link";
        catLink.addEventListener("click", (e) => {
          e.stopPropagation();
          e.preventDefault();
          navigateTo(`${basePath}/category/${cat.id}`);
        });

        catContainer.appendChild(catLink);
        if (idx < post.categories.length - 1) {
          catContainer.appendChild(document.createTextNode(", "));
        }
      });

      wrapper.insertBefore(catContainer, header.nextSibling);
    }

    container.appendChild(wrapper);
  });
}

export function mergePostsFromCategories(categories) {
  const postsMap = new Map();

  (categories || []).forEach((category) => {
    const catId = category.id;
    const catName = category.name;

    (category.posts || []).forEach((post) => {
      const postId = post.id;
      if (!postsMap.has(postId)) {
        postsMap.set(postId, {
          ...post,
          categories: [{ id: catId, name: catName }],
        });
      } else {
        const existing = postsMap.get(postId);
        if (!existing.categories.some((c) => c.id === catId)) {
          existing.categories.push({ id: catId, name: catName });
        }
      }
    });
  });

  return Array.from(postsMap.values());
}
