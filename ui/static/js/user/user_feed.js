const feedURL = "http://localhost:8080/forum/api/feed";

async function loadFeed() {
  try {
    const resp = await fetch(feedURL, { credentials: "include" });
    if (!resp.ok) throw new Error("Failed to load feed");

    const data = await resp.json();
    renderFeed(data.categories || []);
  } catch (err) {
    console.error("Error loading feed:", err);
  }
}

function renderFeed(categories) {
  const container = document.getElementById("forumContainer");
  container.innerHTML = "";

  const posts = mergePostsFromCategories(categories).sort((a, b) => {
    const timeA = new Date(a.created_at).getTime();
    const timeB = new Date(b.created_at).getTime();
    return timeB - timeA; // descending (newest first)
  });

  if (posts.length === 0) {
    container.textContent = "No posts available";
    return;
  }

  const postTpl = document.getElementById("post-template");

  posts.forEach((post) => {
    const postNode = postTpl.content.cloneNode(true);
    const postElement = postNode.querySelector(".post");

    if (post.thumbnail_url) {
      const img = document.createElement("img");
      img.src = post.thumbnail_url;
      img.alt = "Post thumbnail";
      img.className = "post-thumb";
      postElement.insertBefore(img, postElement.firstChild);
    }

    // Make the entire post clickable
    postElement.classList.add("clickable-post");
    postElement.style.cursor = "pointer";
    postElement.addEventListener("click", () => {
      if (post.id) {
        window.location.href = `/user/post?id=${encodeURIComponent(post.id)}`;
      }
    });

    // Username
    postNode.querySelector(".post-header").textContent =
      post.username || post.user_id || "Unknown user";

    // Title & content
    postNode.querySelector(".post-title").textContent = post.title || "";
    postNode.querySelector(".post-content").textContent = post.content || "";

    // Timestamp
    if (post.created_at) {
      postNode.querySelector(".post-time").textContent = new Date(
        post.created_at
      ).toLocaleString();
    }

    // Reactions
    const likes =
      post.reactions?.filter((r) => r.reaction_type === 1).length || 0;
    const dislikes =
      post.reactions?.filter((r) => r.reaction_type === 2).length || 0;
    postNode.querySelector(".like-count").textContent = likes;
    postNode.querySelector(".dislike-count").textContent = dislikes;

    // Comments
    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentContainer = document.createElement("span");
    commentContainer.className = "comment-count";
    commentContainer.innerHTML = `💬 ${commentCount}`;
    postNode
      .querySelector(".like-count")
      .parentNode.appendChild(commentContainer);

    // Categories
    const catContainer = document.createElement("div");
    catContainer.className = "post-categories";

    const postedOnSpan = document.createElement("span");
    postedOnSpan.classList.add("posted-on-text");
    postedOnSpan.textContent = "posted on the ";
    catContainer.appendChild(postedOnSpan);

    post.categories.forEach((cat, idx) => {
      const catLink = document.createElement("a");
      catLink.href = `/user/category?id=${encodeURIComponent(cat.id)}`;
      catLink.textContent = cat.name;
      catLink.classList.add("post-category-link");
      catContainer.appendChild(catLink);

      if (idx < post.categories.length - 1) {
        catContainer.appendChild(document.createTextNode(", "));
      }
    });

    const postTitleEl = postNode.querySelector(".post-title");
    postTitleEl.parentNode.insertBefore(catContainer, postTitleEl);

    container.appendChild(postNode);
  });
}

function mergePostsFromCategories(categories) {
  const postsMap = new Map();

  categories.forEach((category) => {
    const categoryId = category.id || category.ID;
    const categoryName = category.name || category.Name;

    category.posts.forEach((post) => {
      const postId = post.id || post.ID;
      if (!postsMap.has(postId)) {
        postsMap.set(postId, {
          ...post,
          categories: [{ id: categoryId, name: categoryName }],
        });
      } else {
        const existingPost = postsMap.get(postId);
        if (!existingPost.categories.some((c) => c.id === categoryId)) {
          existingPost.categories.push({ id: categoryId, name: categoryName });
        }
      }
    });
  });

  return Array.from(postsMap.values());
}

const logoutLink = document.getElementById("logout-link");

if (logoutLink) {
  logoutLink.addEventListener("click", async (e) => {
    e.preventDefault(); // prevent default anchor navigation
    try {
      const res = await fetch(
        "http://localhost:8080/forum/api/session/logout",
        {
          method: "POST",
          credentials: "include",
        }
      );
      if (res.ok) {
        window.location.href = "/login";
      } else {
        console.error("Logout failed with status:", res.status);
      }
    } catch (err) {
      console.error("Logout error:", err);
    }
  });
}

window.addEventListener("DOMContentLoaded", loadFeed);
