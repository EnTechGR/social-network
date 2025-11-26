import { getCSRF, verifySession } from "../session.js";
import { navigateTo } from "../router.js";

const API_BASE = "http://localhost:8080/forum/api";
const lastReactions = new Map(); // key = `${type}:${id}`, value = 1 (like) or 2 (dislike)

export function renderPost(main, postId) {
  main.innerHTML = `
    <section class="post-page fade-in">
      <div id="postContainer" class="post-container">
        <div class="loader">Loading post...</div>
      </div>
    </section>
  `;

  const postContainer = document.getElementById("postContainer");
  let csrfTokenFromResponse = null;

  async function loadCSRFTokenFromSession() {
    try {
      const resp = await fetch(`${API_BASE}/session/verify`, {
        credentials: "include",
      });
      if (!resp.ok) throw new Error("Session not valid");
      const data = await resp.json();
      return data.csrf_token || data.CSRFToken;
    } catch (err) {
      console.warn("Failed to load CSRF token:", err);
      return null;
    }
  }

  async function loadPost() {
    try {
      const resp = await fetch(`${API_BASE}/feed`, { credentials: "include" });
      if (!resp.ok) throw new Error("Failed to load post");

      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []);
      const post = posts.find((p) => p.id == postId);

      if (!post) {
        postContainer.innerHTML = `<p class="error">Post not found.</p>`;
        return;
      }

      renderSinglePost(post);
    } catch (err) {
      console.error("Error loading post:", err);
      postContainer.innerHTML = `<p class="error">Unable to load post. Please try again later.</p>`;
    }
  }

  function getPostDisplayState(post) {
    const isDeleted = post.title === "" && post.content === "";
    return {
      isDeleted,
      displayTitle: isDeleted ? "This post was deleted" : post.title,
      displayContent: isDeleted ? null : post.content,
    };
  }

  function renderSinglePost(post) {
    const { isDeleted, displayTitle, displayContent } =
      getPostDisplayState(post);
    postContainer.innerHTML = "";

    const postBox = document.createElement("article");
    postBox.className = "post-card-detail";

    const title = document.createElement("h1");
    title.className = "post-title";
    title.textContent = displayTitle;
    postBox.appendChild(title);

    const meta = document.createElement("div");
    meta.className = "post-meta";
    const edited = post.updated_at && post.updated_at !== post.created_at;
    const metaLabel = edited ? " (Edited)" : isDeleted ? " (Deleted)" : "";
    meta.textContent = `By ${post.username || "Unknown"} • ${new Date(
      post.updated_at || post.created_at
    ).toLocaleString()}${metaLabel}`;

    postBox.appendChild(meta);

    if (post.image_url) {
      const image = document.createElement("img");
      image.src = post.image_url;
      image.className = "post-image";
      postBox.appendChild(image);
    }

    if (displayContent) {
      const content = document.createElement("div");
      content.className = "post-content";
      content.textContent = displayContent;
      postBox.appendChild(content);
    }

    /* ========== REACTIONS ========== */
    const reactions = document.createElement("div");
    reactions.className = "post-reactions";

    const reactionsArray = Array.isArray(post.reactions) ? post.reactions : [];
    const likes = reactionsArray.filter((r) => r.reaction_type === 1).length;
    const dislikes = reactionsArray.filter((r) => r.reaction_type === 2).length;

    const likeBtn = document.createElement("button");
    likeBtn.className = "like-btn";
    likeBtn.textContent = `▲ ${likes}`;
    const dislikeBtn = document.createElement("button");
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${dislikes}`;

    likeBtn.addEventListener("click", () =>
      handleReaction(post.id, "post", 1, likeBtn, dislikeBtn)
    );
    dislikeBtn.addEventListener("click", () =>
      handleReaction(post.id, "post", 2, likeBtn, dislikeBtn)
    );

    reactions.append(likeBtn, dislikeBtn);
    postBox.append(reactions);

    /* ========== CATEGORY TAGS ========== */
    const categoryBar = document.createElement("div");
    categoryBar.className = "post-categories";
    categoryBar.innerHTML = `<span class="Posted-on-text">Posted in </span>`;
    post.categories?.forEach((cat, i) => {
      const a = document.createElement("a");
      a.textContent = cat.name;
      a.href = `/user/category/${cat.id}`;
      a.className = "post-category-link";
      a.addEventListener("click", (e) => {
        e.preventDefault();
        navigateTo(`/user/category/${cat.id}`);
      });
      categoryBar.appendChild(a);
      if (i < post.categories.length - 1)
        categoryBar.appendChild(document.createTextNode(", "));
    });
    postBox.append(categoryBar);

    /* ========== COMMENTS SECTION ========== */
    const comments = document.createElement("div");
    comments.className = "comments-section";
    const header = document.createElement("h3");
    header.textContent = "Comments";
    comments.append(header);

    // Add form for comment creation
    const form = document.createElement("form");
    form.className = "comment-form";
    const textarea = document.createElement("textarea");
    textarea.placeholder = "Write your comment...";
    textarea.maxLength = 1000;
    const submitBtn = document.createElement("button");
    submitBtn.textContent = "Post Comment";
    const message = document.createElement("p");
    message.className = "comment-message";

    form.append(textarea, submitBtn, message);
    comments.append(form);

    // Handle comment submission
    form.addEventListener("submit", async (e) => {
      e.preventDefault();
      const content = textarea.value.trim();
      if (!content) {
        message.textContent = "Please enter a comment.";
        message.classList.add("error");
        return;
      }
      await verifySession();
      const csrf = getCSRF();
      if (!csrf) return navigateTo("/login");

      try {
        const res = await fetch(`${API_BASE}/comments/create`, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "X-CSRF-Token": csrf,
          },
          body: JSON.stringify({ post_id: post.id, content }),
        });
        if (!res.ok) throw new Error("Failed to post comment");
        message.textContent = "Comment posted!";
        message.className = "comment-message success";
        textarea.value = "";
        await loadPost();
      } catch (err) {
        message.textContent = err.message;
        message.classList.add("error");
      }
    });

    // Existing comments
    const commentList = document.createElement("div");
    commentList.className = "comment-list";

    if (post.comments?.length) {
      post.comments.forEach((c) => commentList.appendChild(createComment(c)));
    } else {
      commentList.innerHTML = `<p class="no-comments">No comments yet.</p>`;
    }

    comments.append(commentList);
    postBox.append(comments);
    postContainer.append(postBox);
  }

  function createComment(c) {
    const el = document.createElement("div");
    el.className = "comment";

    // Convert to Date objects safely
    const createdDate = new Date(c.created_at);
    const updatedDate = c.updated_at ? new Date(c.updated_at) : createdDate;

    // Extract timestamps
    const createdMs = createdDate.getTime();
    const updatedMs = updatedDate.getTime();

    // If any timestamp is invalid, avoid "Invalid Date"
    const safeDate = isFinite(updatedMs) ? updatedDate : createdDate;

    // Determine if edited:
    // Only if backend actually includes updated_at
    // AND both dates valid
    // AND difference is significant (> 1 second)
    const isEdited =
      c.updated_at &&
      isFinite(createdMs) &&
      isFinite(updatedMs) &&
      Math.abs(updatedMs - createdMs) > 1000;

    // Build header HTML
    const header = document.createElement("div");
    header.className = "comment-header";
    header.innerHTML = `<strong>${
      c.username || "Anonymous"
    }</strong> • ${safeDate.toLocaleString()}${isEdited ? " (Edited)" : ""}`;

    const body = document.createElement("p");
    body.textContent = c.content || "This comment was deleted.";
    body.className = "comment-body";

    // Reaction buttons
    const reacts = document.createElement("div");
    reacts.className = "comment-reactions";
    const likes = c.reactions?.filter((r) => r.reaction_type === 1).length || 0;
    const dislikes =
      c.reactions?.filter((r) => r.reaction_type === 2).length || 0;

    const likeBtn = document.createElement("button");
    likeBtn.type = "button";
    likeBtn.className = "like-btn";
    likeBtn.textContent = `▲ ${likes}`;

    const dislikeBtn = document.createElement("button");
    dislikeBtn.type = "button";
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${dislikes}`;

    likeBtn.addEventListener("click", () =>
      handleReaction(c.id, "comment", 1, likeBtn, dislikeBtn)
    );
    dislikeBtn.addEventListener("click", () =>
      handleReaction(c.id, "comment", 2, likeBtn, dislikeBtn)
    );

    reacts.append(likeBtn, dislikeBtn);
    el.append(header, body, reacts);

    return el;
  }

  async function handleReaction(
    targetId,
    type,
    reactionType,
    likeBtn,
    dislikeBtn
  ) {
    const key = `${type}:${targetId}`;
    const prev = lastReactions.get(key);
    const remove = prev === reactionType;
    const finalType = remove ? 3 : reactionType;

    await verifySession();
    const csrf = getCSRF();
    if (!csrf) return navigateTo("/login");

    likeBtn.disabled = true;
    dislikeBtn.disabled = true;

    try {
      const res = await fetch(`${API_BASE}/react`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        body: JSON.stringify({
          target_id: targetId,
          target_type: type,
          reaction_type: finalType,
        }),
      });

      if (!res.ok) throw new Error("Failed to react");

      const reactions = await res.json();
      const likes = reactions.filter((r) => r.reaction_type === 1).length;
      const dislikes = reactions.filter((r) => r.reaction_type === 2).length;

      likeBtn.textContent = `▲ ${likes}`;
      dislikeBtn.textContent = `▼ ${dislikes}`;

      if (remove) lastReactions.delete(key);
      else lastReactions.set(key, reactionType);
    } catch (err) {
      console.error("Reaction error:", err);
      alert("Error: " + err.message);
    } finally {
      likeBtn.disabled = false;
      dislikeBtn.disabled = false;
    }
  }

  function mergePostsFromCategories(categories) {
    const map = new Map();
    for (const cat of categories) {
      for (const post of cat.posts || []) {
        if (!map.has(post.id)) {
          map.set(post.id, {
            ...post,
            categories: [{ id: cat.id, name: cat.name }],
          });
        } else {
          map.get(post.id).categories.push({ id: cat.id, name: cat.name });
        }
      }
    }
    return Array.from(map.values());
  }

  loadPost();
}
