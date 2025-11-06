import { verifySession, getCSRF } from "../session.js";
import { navigateTo } from "../router.js";

let csrfTokenFromResponse = null;
let currentUserId = null;

// Main entry point called from router
export function renderEditPost(main, postId) {
  main.innerHTML = `
    <section class="edit-post-page">
      <header class="category-header">
        <h2 class="category-title">Edit Post</h2>
        <p class="category-description">
          Update your thoughts, refine your ideas, or expand your story.
        </p>
      </header>

      <div id="postContainer" class="post-card-detail">
        <div class="loader">Loading post...</div>
      </div>
    </section>
  `;

  const postContainer = document.getElementById("postContainer");

  //
  // =====================================================
  // Everything below is your original logic, ADAPTED
  // =====================================================
  //

  // Query param already provided by router as postId

  async function loadPost() {
    if (!postId) {
      postContainer.textContent = "Post ID missing.";
      return;
    }

    try {
      const resp = await fetch("http://localhost:8080/forum/api/feed", {
        credentials: "include",
      });
      if (!resp.ok) throw new Error("Failed to load post");

      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []);
      const post = posts.find((p) => p.id === postId);

      if (!post) {
        postContainer.textContent = "Post not found.";
        return;
      }

      renderSinglePostWithEdit(post);
    } catch (err) {
      console.error(err);
      postContainer.textContent = "Error loading post.";
    }
  }

  function getPostDisplayState(post) {
    let isDeleted = false;
    let displayTitle = post.title;
    let displayContent = post.content;

    if (post.title === "" && post.content === "") {
      displayTitle = "This post was deleted";
      displayContent = null;
      isDeleted = true;
    }

    return { isDeleted, displayTitle, displayContent };
  }

  function renderSinglePostWithEdit(post) {
    postContainer.innerHTML = "";

    const { isDeleted, displayTitle, displayContent } =
      getPostDisplayState(post);

    const postBox = document.createElement("div");
    postBox.className = "post";

    // TITLE
    const title = document.createElement("h1");
    title.className = isDeleted ? "deleted-title" : "post-title";
    title.textContent = displayTitle;
    postBox.appendChild(title);

    const titleEditBtn = document.createElement("button");
    titleEditBtn.textContent = "Edit Title";
    titleEditBtn.className = "edit-btn";
    titleEditBtn.onclick = () => showEditTitle(post.title);
    if (isDeleted) titleEditBtn.style.display = "none";

    // META
    const meta = document.createElement("div");
    meta.className = "post-meta";

    let metaDate;
    let edited = false;

    if (post.updated_at && post.updated_at !== post.created_at) {
      metaDate = new Date(post.updated_at).toLocaleString();
      edited = true;
    } else {
      metaDate = new Date(post.created_at).toLocaleString();
    }

    meta.textContent =
      `By ${post.username} on ${metaDate}` +
      (isDeleted ? " (Deleted)" : edited ? " (Edited)" : "");

    // IMAGE
    let imageEl = null;
    if (post.image_url) {
      imageEl = document.createElement("img");
      imageEl.className = "post-image";
      imageEl.src = post.image_url;
    }

    const imageEditBtn = document.createElement("button");
    imageEditBtn.textContent = "Edit Image";
    imageEditBtn.className = "edit-btn";
    imageEditBtn.onclick = () => showEditImage(imageEditBtn);
    if (isDeleted || !post.image_url) imageEditBtn.style.display = "none";

    // CONTENT
    const contentWrapper = document.createElement("div");
    contentWrapper.className = "post-content-card";

    const content = document.createElement("div");
    content.className = isDeleted ? "deleted-content" : "post-content";
    content.textContent = displayContent;
    if (!isDeleted) contentWrapper.appendChild(content);

    const contentEditBtn = document.createElement("button");
    contentEditBtn.textContent = "Edit Content";
    contentEditBtn.className = "edit-btn";
    contentEditBtn.onclick = () => showEditContent(post.content);
    if (isDeleted) contentEditBtn.style.display = "none";

    // DELETE BUTTON
    const deleteBtn = document.createElement("button");
    deleteBtn.textContent = "Delete";
    deleteBtn.className = "delete-btn";
    deleteBtn.onclick = deletePost;
    if (isDeleted) deleteBtn.style.display = "none";

    // REACTIONS
    const reactions = document.createElement("div");
    reactions.className = "post-reactions";

    const reactionsArr = Array.isArray(post.reactions) ? post.reactions : [];

    const likes = reactionsArr.filter((r) => r.reaction_type === 1).length;
    const dislikes = reactionsArr.filter((r) => r.reaction_type === 2).length;

    const likeBtn = document.createElement("button");
    likeBtn.className = "like-btn";
    likeBtn.textContent = `▲ ${likes}`;
    likeBtn.disabled = true;

    const dislikeBtn = document.createElement("button");
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${dislikes}`;
    dislikeBtn.disabled = true;

    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);

    const commentCounter = document.createElement("span");
    commentCounter.className = "comment-count";
    commentCounter.textContent = `💬 ${commentCount}`;

    reactions.append(likeBtn, dislikeBtn, commentCounter);

    // CATEGORIES
    const categoryEl = document.createElement("div");
    categoryEl.className = "post-categories";
    categoryEl.innerHTML = `<span class="Posted-on-text">Posted in </span>`;

    post.categories?.forEach((cat, i) => {
      const link = document.createElement("a");
      link.href = `/user/category/${cat.id}`;
      link.textContent = cat.name;
      link.className = "post-category-link";
      categoryEl.appendChild(link);
      if (i < post.categories.length - 1)
        categoryEl.appendChild(document.createTextNode(", "));
    });

    // COMMENTS
    const commentSection = document.createElement("div");
    commentSection.className = "comments-section";

    const commentHeader = document.createElement("h3");
    commentHeader.textContent = "Comments";
    commentSection.appendChild(commentHeader);

    // COMMENT FORM
    const commentFormContainer = document.createElement("div");
    commentFormContainer.className = "comment-form-container";

    const commentForm = document.createElement("form");
    commentForm.className = "comment-form";

    const textarea = document.createElement("textarea");
    textarea.className = "comment-textarea";
    textarea.placeholder = "Write your comment...";
    textarea.maxLength = 1000;

    const charCount = document.createElement("div");
    charCount.className = "comment-char-count";
    charCount.textContent = "0 / 1000";

    textarea.addEventListener("input", () => {
      charCount.textContent = `${textarea.value.length} / 1000`;
      if (textarea.value.length > 1000)
        textarea.value = textarea.value.slice(0, 1000);
    });

    const submitCommentBtn = document.createElement("button");
    submitCommentBtn.className = "submit-comment-btn";
    submitCommentBtn.textContent = "Submit Comment";

    commentForm.append(textarea, charCount, submitCommentBtn);
    commentFormContainer.appendChild(commentForm);

    // SUBMIT COMMENT (same logic as old)
    commentForm.addEventListener("submit", async (e) => {
      e.preventDefault();

      const text = textarea.value.trim();
      if (!text) return;

      const csrf = await getCSRF();

      await fetch("http://localhost:8080/forum/api/comments/create", {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        body: JSON.stringify({ post_id: post.id, content: text }),
      });

      textarea.value = "";
      charCount.textContent = "0 / 1000";

      loadPost();
    });

    // APPEND ELEMENTS IN BOOKISH ORDER
    postBox.appendChild(meta);
    if (post.image_url) postBox.appendChild(imageEl);
    if (post.image_url) postBox.appendChild(imageEditBtn);
    if (!isDeleted) postBox.appendChild(contentWrapper);
    if (!isDeleted) postBox.appendChild(contentEditBtn);
    if (!isDeleted) postBox.appendChild(titleEditBtn);
    postBox.appendChild(reactions);
    postBox.appendChild(categoryEl);
    if (!isDeleted) postBox.appendChild(deleteBtn);
    if (!isDeleted) postBox.appendChild(commentFormContainer);

    if (post.comments?.length) {
      post.comments.forEach((comment) =>
        commentSection.appendChild(createCommentElement(comment, isDeleted))
      );
    } else {
      const empty = document.createElement("p");
      empty.className = "no-comments";
      empty.textContent = "No comments yet.";
      commentSection.appendChild(empty);
    }

    postBox.appendChild(commentSection);

    postContainer.appendChild(postBox);
  }

  //
  // --- COMMENT ELEMENT BUILDER (original logic unchanged)
  //
  function createCommentElement(comment, isPostDeleted) {
    const el = document.createElement("div");
    el.className = "comment";

    const user = document.createElement("strong");
    user.textContent = comment.username || "Anonymous";

    const time = document.createElement("time");
    if (comment.content === "") {
      time.textContent = ` (${new Date(
        comment.updated_at
      ).toLocaleString()}) (Deleted)`;
    } else if (
      comment.updated_at &&
      comment.updated_at !== comment.created_at
    ) {
      time.textContent = ` (${new Date(
        comment.updated_at
      ).toLocaleString()}) (Edited)`;
    } else {
      time.textContent = ` (${new Date(comment.created_at).toLocaleString()})`;
    }

    const content = document.createElement("div");
    content.className = "comment-content";
    content.textContent =
      comment.content === "" ? "This comment was deleted" : comment.content;

    const reactions = document.createElement("div");
    reactions.className = "comment-reactions";

    const arr = Array.isArray(comment.reactions) ? comment.reactions : [];

    const likeBtn = document.createElement("button");
    likeBtn.className = "like-btn";
    likeBtn.textContent = `▲ ${
      arr.filter((r) => r.reaction_type === 1).length
    }`;
    likeBtn.disabled = true;

    const dislikeBtn = document.createElement("button");
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${
      arr.filter((r) => r.reaction_type === 2).length
    }`;
    dislikeBtn.disabled = true;

    reactions.append(likeBtn, dislikeBtn);

    el.append(user, time, content, reactions);

    return el;
  }

  //
  // --- SHOW EDIT TITLE (original)
  //
  function showEditTitle(current) {
    const title = document.querySelector(".post-title, .deleted-title");
    title.innerHTML = `
      <input id="titleInput" value="${current ?? ""}" class="edit-text-input" />
      <button id="saveTitleBtn" class="small-btn">Save</button>
      <button id="cancelTitleBtn" class="small-btn muted">Cancel</button>
    `;

    document.getElementById("saveTitleBtn").onclick = saveTitle;
    document.getElementById("cancelTitleBtn").onclick = loadPost;
  }

  async function saveTitle() {
    const newTitle = document.getElementById("titleInput").value;
    const csrf = await getCSRF();

    await fetch(`http://localhost:8080/forum/api/posts/edit-title/${postId}`, {
      method: "PUT",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrf,
      },
      body: JSON.stringify({ title: newTitle }),
    });
    loadPost();
  }

  //
  // --- SHOW EDIT CONTENT
  //
  function showEditContent(current) {
    const content = document.querySelector(".post-content, .deleted-content");
    content.innerHTML = `
      <textarea id="contentInput" class="edit-textarea">${
        current ?? ""
      }</textarea>
      <button id="saveContentBtn" class="small-btn">Save</button>
      <button id="cancelContentBtn" class="small-btn muted">Cancel</button>
    `;

    document.getElementById("saveContentBtn").onclick = saveContent;
    document.getElementById("cancelContentBtn").onclick = loadPost;
  }

  async function saveContent() {
    const newContent = document.getElementById("contentInput").value;
    const csrf = await getCSRF();

    await fetch(
      `http://localhost:8080/forum/api/posts/edit-content/${postId}`,
      {
        method: "PUT",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        body: JSON.stringify({ content: newContent }),
      }
    );
    loadPost();
  }

  //
  // --- SHOW EDIT IMAGE (original logic preserved)
  //
  function showEditImage(btn) {
    const image = document.querySelector(".post-image");
    if (!image) return;

    btn.disabled = true;
    btn.style.opacity = "0.4";

    const box = document.createElement("div");
    box.className = "image-upload-interface";
    box.innerHTML = `
      <input type="file" id="imageInput" class="hidden-file" accept="image/*" />
      <button id="chooseImgBtn" class="small-btn">Choose Image</button>
      <button id="cancelImgBtn" class="small-btn muted hidden">Cancel</button>
      <img id="previewImg" class="hidden preview-img" />
      <button id="uploadImgBtn" class="small-btn accent hidden" disabled>Upload</button>
      <div id="imgError" class="error-message"></div>
    `;

    image.after(box);

    const input = box.querySelector("#imageInput");
    const chooseBtn = box.querySelector("#chooseImgBtn");
    const cancelBtn = box.querySelector("#cancelImgBtn");
    const preview = box.querySelector("#previewImg");
    const uploadBtn = box.querySelector("#uploadImgBtn");
    const err = box.querySelector("#imgError");

    chooseBtn.onclick = () => input.click();
    cancelBtn.onclick = () => {
      box.remove();
      btn.disabled = false;
      btn.style.opacity = "1";
    };

    input.onchange = () => {
      const file = input.files[0];
      if (!file) return;

      if (!["image/jpeg", "image/png", "image/gif"].includes(file.type)) {
        err.textContent = "Unsupported file type.";
        input.value = "";
        return;
      }
      if (file.size > 20 * 1024 * 1024) {
        err.textContent = "Image exceeds 20 MB.";
        input.value = "";
        return;
      }

      const reader = new FileReader();
      reader.onload = (e) => {
        preview.src = e.target.result;
        preview.classList.remove("hidden");
        uploadBtn.classList.remove("hidden");
        cancelBtn.classList.remove("hidden");
        uploadBtn.disabled = false;
      };
      reader.readAsDataURL(file);
    };

    uploadBtn.onclick = async () => {
      const file = input.files[0];
      if (!file) return;

      const csrf = await getCSRF();

      const formData = new FormData();
      formData.append("post_id", postId);
      formData.append("image", file);

      await fetch("http://localhost:8080/forum/api/images/upload", {
        method: "POST",
        credentials: "include",
        headers: { "X-CSRF-Token": csrf },
        body: formData,
      });

      box.remove();
      btn.disabled = false;
      btn.style.opacity = "1";

      loadPost();
    };
  }

  //
  // --- DELETE POST
  //
  async function deletePost() {
    if (!confirm("Are you sure you want to delete this post?")) return;

    const csrf = await getCSRF();

    await fetch(`http://localhost:8080/forum/api/posts/delete/${postId}`, {
      method: "DELETE",
      credentials: "include",
      headers: { "X-CSRF-Token": csrf },
    });

    await fetch(`http://localhost:8080/forum/api/images/delete/${postId}`, {
      method: "DELETE",
      credentials: "include",
      headers: { "X-CSRF-Token": csrf },
    });

    navigateTo("/user/my-activity/my-posts");
  }

  //
  // --- MERGE POSTS FROM FEED
  //
  function mergePostsFromCategories(categories) {
    const map = new Map();

    categories.forEach((cat) => {
      cat.posts.forEach((p) => {
        if (!map.has(p.id)) {
          map.set(p.id, {
            ...p,
            categories: [{ id: cat.id, name: cat.name }],
          });
        } else {
          map.get(p.id).categories.push({
            id: cat.id,
            name: cat.name,
          });
        }
      });
    });

    return Array.from(map.values());
  }

  // ------------------------------------------------------
  // LOAD SESSION + CSRF THEN LOAD POST
  // ------------------------------------------------------
  (async () => {
    const user = await verifySession();

    if (!user) {
      alert("Your session expired. Please log in again.");
      window.location.href = "/login";
      return;
    }

    csrfTokenFromResponse = getCSRF();
    currentUserId = user.id;

    loadPost();
  })();
}
