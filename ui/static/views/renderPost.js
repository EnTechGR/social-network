import { getCSRF, verifySession } from "../session.js";
import { navigateTo } from "../router.js";

const API_BASE = "http://localhost:8080/forum/api";
const lastReactions = new Map(); // key = `${type}:${id}`, value = 1 (like) or 2 (dislike)

let csrfTokenFromResponse = null;
let currentUserId = null;

// Helper: merge posts from categories (reused from old code)
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

// Helper to determine deleted post display state (reused from old code)
function getPostDisplayState(post) {
  const isDeleted = post.title === "" && post.content === "";
  return {
    isDeleted,
    displayTitle: isDeleted ? "This post was deleted" : post.title,
    displayContent: isDeleted ? null : post.content,
  };
}

// Helper to fetch current user ID (reused from old code)
async function fetchCurrentUserId() {
  try {
    const resp = await fetch(`${API_BASE}/session/verify`, {
      credentials: "include",
    });
    if (!resp.ok) return null;
    const data = await resp.json();
    return data.user?.id || (data.user && data.user.ID) || null;
  } catch (err) {
    return null;
  }
}


// --- PATCH: Change to async function to use await for auth check ---
export async function renderPost(main, postId) { 
  main.innerHTML = `
    <section class="post-page fade-in">
      <div id="postContainer" class="post-container">
        <div class="loader">Loading post...</div>
      </div>
    </section>
  `;

  const postContainer = document.getElementById("postContainer");
  let isUserAuthenticated = false; // Flag to enable/disable reaction/comment forms

  // ----------------------------------------------------
  // 🔑 PATCH 1: Explicitly await Auth/CSRF Load for SPA navigation
  // This ensures currentUserId is set before loadPost calls renderSinglePost.
  if (!csrfTokenFromResponse) {
    csrfTokenFromResponse = await loadCSRFTokenFromSession();
  }
  
  if (csrfTokenFromResponse) {
    // If we have a CSRF token, the user is authenticated
    isUserAuthenticated = true;
  } else {
    isUserAuthenticated = false;
  }
  // ----------------------------------------------------

  async function loadCSRFTokenFromSession() {
    try {
      const resp = await fetch(`${API_BASE}/session/verify`, {
        credentials: "include",
      });
      if (!resp.ok) {
        isUserAuthenticated = false;
        throw new Error("Session not valid");
      }
      isUserAuthenticated = true;
      const data = await resp.json();
      currentUserId = data.user?.id || (data.user && data.user.ID) || null;
      return data.csrf_token || data.CSRFToken;
    } catch (err) {
      isUserAuthenticated = false;
      console.warn("Failed to load CSRF token/user info:", err);
      return null;
    }
  }

  // Helper to get auth headers (including CSRF) (reused from old code)
  async function getAuthHeaders(skipContentType) {
    const headers = {};
    if (!skipContentType) headers['Content-Type'] = 'application/json';
    if (csrfTokenFromResponse) {
      headers['X-CSRF-Token'] = csrfTokenFromResponse;
    } else {
      // Attempt to load if not already loaded, though loadPost should ensure it's loaded
      csrfTokenFromResponse = await loadCSRFTokenFromSession();
      if (csrfTokenFromResponse) {
        headers['X-CSRF-Token'] = csrfTokenFromResponse;
      }
    }
    return headers;
  }

  // --- EDIT/DELETE LOGIC (Integrated from old code) ---

  async function deletePost() {
    if (!window.confirm('Are you sure you want to delete this post?')) return;
    try {
      // Delete the post (soft-delete)
      const deletePostResp = await fetch(`${API_BASE}/posts/delete/${postId}`, {
        method: 'DELETE',
        headers: await getAuthHeaders(),
        credentials: 'include'
      });

      if (deletePostResp.ok) {
        // Use SPA navigation/reload logic
        await loadPost();
        // Alternatively, navigate back to the feed:
        // navigateTo('/');
      } else {
        throw new Error('Post deletion failed.');
      }

    } catch (error) {
      console.error('Error deleting post:', error);
      window.alert('Failed to delete post. Please try again.');
    }
  }

  function showEditTitle(currentTitle) {
    // Find the title element and replace it with edit form
    const titleElement = postContainer.querySelector('.post-title, .deleted-title');
    const titleEditBtn = postContainer.querySelector('.edit-title-btn');

    if (titleElement) {
      titleElement.innerHTML = `
        <input id="titleInput" value="${currentTitle ?? ''}" style="width:60%;-webkit-text-fill-color:black"/> 
        <button type="button" id="saveTitleBtn" style="-webkit-text-fill-color:black">Save</button> 
        <button type="button" id="cancelTitleBtn" style="-webkit-text-fill-color:black">Cancel</button>
      `;
      if (titleEditBtn) titleEditBtn.style.display = 'none';

      document.getElementById('saveTitleBtn').onclick = saveTitle;
      document.getElementById('cancelTitleBtn').onclick = loadPost;
    }
  }

  async function saveTitle() {
    const newTitle = document.getElementById('titleInput')?.value;
    if (!newTitle) return loadPost(); // Cancel if no input

    try {
      await fetch(`${API_BASE}/posts/edit-title/${postId}`, {
        method: 'PUT',
        headers: await getAuthHeaders(),
        body: JSON.stringify({ title: newTitle }),
        credentials: 'include'
      });
    } catch (error) {
      console.error('Error saving title:', error);
      window.alert('Failed to update title.');
    }
    loadPost();
  }

  function showEditContent(currentContent) {
    // Find the content element and replace it with edit form
    const contentElement = postContainer.querySelector('.post-content, .deleted-content');
    const contentEditBtn = postContainer.querySelector('.edit-content-btn');
    if (contentElement) {
      contentElement.innerHTML = `
        <textarea id="contentInput" style="width:90%; min-height: 150px">${currentContent ?? ''}</textarea><br/>
        <button type="button" id="saveContentBtn">Save</button> 
        <button type="button" id="cancelContentBtn">Cancel</button>
      `;
      if (contentEditBtn) contentEditBtn.style.display = 'none';

      document.getElementById('saveContentBtn').onclick = saveContent;
      document.getElementById('cancelContentBtn').onclick = loadPost;
    }
  }

  async function saveContent() {
    const newContent = document.getElementById('contentInput')?.value;
    if (newContent === undefined) return loadPost();

    try {
      await fetch(`${API_BASE}/posts/edit-content/${postId}`, {
        method: 'PUT',
        headers: await getAuthHeaders(),
        body: JSON.stringify({ content: newContent }),
        credentials: 'include'
      });
    } catch (error) {
      console.error('Error saving content:', error);
      window.alert('Failed to update content.');
    }
    loadPost();
  }

  function showEditImage(imageEditBtn) {
    // ... (Image editing logic integrated here, using postContainer for lookups) ...

    // Find the image element and add upload interface right after it
    const imageElement = postContainer.querySelector('.post-image');
    if (imageElement) {
        // Disable the edit image button to prevent multiple upload interfaces
        imageEditBtn.disabled = true;
        imageEditBtn.style.opacity = '0.5';
        // Create upload interface
        const uploadContainer = document.createElement('div');
        uploadContainer.className = 'image-upload-interface';
        uploadContainer.innerHTML = `
            <div class="upload-controls">
                <button type="button" id="addImageBtn" class="add-image-btn">Choose Image</button>
                <button type="button" id="cancelImageBtn" class="cancel-image-btn hidden">Cancel</button>
            </div>
            <input type="file" id="imageInput" accept="image/*" style="display: none;" />
            <div id="imageStatus" class="image-status hidden"></div>
            <div id="imageError" class="image-error"></div>
            <img id="imagePreview" class="image-preview hidden" alt="Image preview" style="max-width: 150px; max-height: 150px; object-fit: cover;" />
            <button type="button" id="uploadImageBtn" class="upload-btn hidden" disabled>Upload Image</button>
        `;

        // Insert right after the image element
        imageElement.parentNode.insertBefore(uploadContainer, imageElement.nextSibling);

        const imageInput = document.getElementById('imageInput');
        const addImageBtn = document.getElementById('addImageBtn');
        const cancelImageBtn = document.getElementById('cancelImageBtn');
        const uploadImageBtn = document.getElementById('uploadImageBtn');
        const imageStatus = document.getElementById('imageStatus');
        const imageError = document.getElementById('imageError');
        const imagePreview = document.getElementById('imagePreview');

        function resetImageSelection() {
            imageInput.value = "";
            imageStatus.textContent = "";
            imageStatus.classList.add("hidden");
            imageStatus.classList.remove("status-valid", "status-error");
            imageError.textContent = "";
            cancelImageBtn.classList.add("hidden");
            uploadImageBtn.classList.add("hidden");
            addImageBtn.disabled = false;
            uploadImageBtn.disabled = true;
            imagePreview.src = "";
            imagePreview.classList.add("hidden");
        }

        function validateSelectedImage() {
            imageError.textContent = "";
            const file = imageInput.files[0];
            if (!file) {
                return true;
            }
            
            const allowed = ["image/jpeg", "image/png", "image/gif"];
            if (!allowed.includes(file.type)) {
                imageStatus.textContent = file.name;
                imageStatus.classList.remove("hidden", "status-valid");
                imageStatus.classList.add("status-error");
                imageError.textContent = "Unsupported image type. Only jpeg, png, gif";
                imageInput.value = "";
                imagePreview.src = "";
                imagePreview.classList.add("hidden");
                cancelImageBtn.classList.remove("hidden");
                uploadImageBtn.classList.add("hidden");
                uploadImageBtn.disabled = true;
                return false;
            }
            
            if (file.size > 20 * 1024 * 1024) {
                imageStatus.textContent = file.name;
                imageStatus.classList.remove("hidden", "status-valid");
                imageStatus.classList.add("status-error");
                imageError.textContent = "Image exceeds 20 MB limit";
                imageInput.value = "";
                imagePreview.src = "";
                imagePreview.classList.add("hidden");
                cancelImageBtn.classList.remove("hidden");
                uploadImageBtn.classList.add("hidden");
                uploadImageBtn.disabled = true;
                return false;
            }

            imageStatus.textContent = file.name;
            imageStatus.classList.remove("hidden", "status-error");
            imageStatus.classList.add("status-valid");
            cancelImageBtn.classList.remove("hidden");
            uploadImageBtn.classList.remove("hidden");
            addImageBtn.disabled = true;
            uploadImageBtn.disabled = false;
            
            const reader = new FileReader();
            reader.onload = (e) => {
                imagePreview.src = e.target.result;
                imagePreview.classList.remove("hidden");
            };
            reader.readAsDataURL(file);
            return true;
        }

        // Event listeners
        addImageBtn.addEventListener("click", () => imageInput.click());
        cancelImageBtn.addEventListener("click", () => {
            resetImageSelection();
            uploadContainer.remove();
            // Re-enable the edit image button
            imageEditBtn.disabled = false;
            imageEditBtn.style.opacity = '1';
        });

        imageInput.addEventListener("change", () => {
            validateSelectedImage();
        });

        uploadImageBtn.addEventListener("click", async () => {
            if (!imageInput.files[0]) return;
            
            uploadImageBtn.disabled = true;
            uploadImageBtn.textContent = "Uploading...";
            
            try {
                const formData = new FormData();
                formData.append('post_id', postId);
                formData.append('image', imageInput.files[0]);
                
                const resp = await fetch(`${API_BASE}/images/upload`, {
                    method: 'POST',
                    headers: await getAuthHeaders(true), // Skip Content-Type for FormData
                    body: formData,
                    credentials: 'include'
                });
                
                if (!resp.ok) {
                    const errorData = await resp.json().catch(() => ({}));
                    throw new Error(errorData.message || 'Upload failed');
                }
                
                // Remove upload interface and reload the post to show the new image
                uploadContainer.remove();
                // Re-enable the edit image button
                imageEditBtn.disabled = false;
                imageEditBtn.style.opacity = '1';
                loadPost();
            } catch (error) {
                console.error('Image upload failed:', error);
                imageError.textContent = `Upload failed: ${error.message}`;
                uploadImageBtn.disabled = false;
                uploadImageBtn.textContent = "Upload Image";
                // Re-enable the edit image button on error
                imageEditBtn.disabled = false;
                imageEditBtn.style.opacity = '1';
            }
        });
    }
  }
  // --- END EDIT/DELETE LOGIC ---


  async function loadPost() {

    try {
      const resp = await fetch(`${API_BASE}/feed`, { credentials: "include" });
      if (!resp.ok) throw new Error("Failed to load feed");

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

    // --- ADD EDIT TITLE BUTTON ---
    if (isUserAuthenticated && post.user_id === currentUserId && !isDeleted) {
      const titleEditBtn = document.createElement("button");
      titleEditBtn.textContent = "Edit Title";
      titleEditBtn.className = "edit-btn edit-title-btn";
      titleEditBtn.addEventListener("click", () => showEditTitle(post.title));
      postBox.appendChild(titleEditBtn);
    }
    // --- END EDIT TITLE BUTTON ---

    const meta = document.createElement("div");
    meta.className = "post-meta";
    const edited = post.updated_at && post.updated_at !== post.created_at;
    const metaLabel = edited ? " (Edited)" : isDeleted ? " (Deleted)" : "";
    meta.textContent = `By ${post.username || "Unknown"} • ${new Date(
      post.updated_at || post.created_at
    ).toLocaleString()}${metaLabel}`;
    postBox.appendChild(meta); // FIX: Appending the meta element

    let imageEl = null;
    if (post.image_url) {
      const image = document.createElement("img");
      image.src = post.image_url;
      image.className = "post-image";
      imageEl = image;
      postBox.appendChild(image);
    }

    // --- ADD EDIT IMAGE BUTTON ---
    if (isUserAuthenticated && post.user_id === currentUserId && !isDeleted && imageEl) {
      const imageEditBtn = document.createElement("button");
      imageEditBtn.textContent = "Edit Image";
      imageEditBtn.className = "edit-btn edit-image-btn";
      imageEditBtn.addEventListener("click", () => showEditImage(imageEditBtn));
      postBox.appendChild(imageEditBtn);
    }
    // --- END EDIT IMAGE BUTTON ---

    if (displayContent) {
      const content = document.createElement("div");
      content.className = "post-content";
      content.textContent = displayContent;
      postBox.appendChild(content);

      // --- ADD EDIT CONTENT BUTTON ---
      if (isUserAuthenticated && post.user_id === currentUserId) {
        const contentEditBtn = document.createElement("button");
        contentEditBtn.textContent = "Edit Content";
        contentEditBtn.className = "edit-btn edit-content-btn";
        contentEditBtn.addEventListener("click", () => showEditContent(post.content));
        postBox.appendChild(contentEditBtn);
      }
      // --- END EDIT CONTENT BUTTON ---
    }

    /* ========== POST DELETE CONTROL ========== */
    if (isUserAuthenticated && post.user_id === currentUserId && !isDeleted) {
      const deleteBtn = document.createElement('button');
      deleteBtn.textContent = 'Delete Post';
      deleteBtn.className = 'delete-btn';
      deleteBtn.addEventListener('click', deletePost);
      postBox.appendChild(deleteBtn);
    }
    /* ========== END POST DELETE CONTROL ========== */


    /* ========== REACTIONS (Unchanged) ========== */
    const reactions = document.createElement("div");
    reactions.className = "post-reactions";

    const reactionsArray = Array.isArray(post.reactions) ? post.reactions : [];
    const likes = reactionsArray.filter((r) => r.reaction_type === 1).length;
    const dislikes = reactionsArray.filter((r) => r.reaction_type === 2).length;

    const likeBtn = document.createElement("button");
    likeBtn.className = "like-btn";
    likeBtn.textContent = `▲ ${likes}`;
    likeBtn.disabled = !isUserAuthenticated;
    
    const dislikeBtn = document.createElement("button");
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${dislikes}`;
    dislikeBtn.disabled = !isUserAuthenticated;

    likeBtn.addEventListener("click", () =>
      handleReaction(post.id, "post", 1, likeBtn, dislikeBtn)
    );
    dislikeBtn.addEventListener("click", () =>
      handleReaction(post.id, "post", 2, likeBtn, dislikeBtn)
    );

    reactions.append(likeBtn, dislikeBtn);
    postBox.append(reactions);

    /* ========== CATEGORY TAGS (Unchanged) ========== */
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

    /* ========== COMMENTS SECTION (Refactored) ========== */
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
    if (isUserAuthenticated) {
      comments.append(form);
    } else {
      comments.append(document.createTextNode("Log in to post a comment."));
    }

    // Handle comment submission
    if (isUserAuthenticated) {
      form.addEventListener("submit", async (e) => {
        e.preventDefault();
        const content = textarea.value.trim();
        if (!content) {
          message.textContent = "Please enter a comment.";
          message.classList.add("error");
          return;
        }
        await verifySession(); // Use imported verifySession
        const csrf = getCSRF(); // Use imported getCSRF
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
    }

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

  // Helper function to create comment element (integrated edit/delete logic)
  function createComment(c) {
    const el = document.createElement("div");
    el.className = "comment";

    const createdDate = new Date(c.created_at);
    const updatedDate = c.updated_at ? new Date(c.updated_at) : createdDate;
    const createdMs = createdDate.getTime();
    const updatedMs = updatedDate.getTime();
    const safeDate = isFinite(updatedMs) ? updatedDate : createdDate;
    const isEdited =
      c.updated_at &&
      isFinite(createdMs) &&
      isFinite(updatedMs) &&
      Math.abs(updatedMs - createdMs) > 1000;

    // Header with user and timestamp
    const header = document.createElement("div");
    header.className = "comment-header";
    header.innerHTML = `<strong>${
      c.username || "Anonymous"
    }</strong> • ${safeDate.toLocaleString()}${
      c.content === "" ? " (Deleted)" : isEdited ? " (Edited)" : ""
    }`;

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
    likeBtn.disabled = !isUserAuthenticated;

    const dislikeBtn = document.createElement("button");
    dislikeBtn.type = "button";
    dislikeBtn.className = "dislike-btn";
    dislikeBtn.textContent = `▼ ${dislikes}`;
    dislikeBtn.disabled = !isUserAuthenticated;

    likeBtn.addEventListener("click", () =>
      handleReaction(c.id, "comment", 1, likeBtn, dislikeBtn)
    );
    dislikeBtn.addEventListener("click", () =>
      handleReaction(c.id, "comment", 2, likeBtn, dislikeBtn)
    );

    reacts.append(likeBtn, dislikeBtn);

    // --- COMMENT EDIT/DELETE CONTROLS ---
    if (isUserAuthenticated && c.user_id === currentUserId && c.content !== "") {
      const editBtn = document.createElement('button');
      editBtn.textContent = 'Edit';
      editBtn.className = 'edit-comment-btn';
      const deleteBtn = document.createElement('button');
      deleteBtn.textContent = 'Delete';
      deleteBtn.className = 'delete-comment-btn';

      // Attach Edit handler
      editBtn.addEventListener('click', () => {
        const textarea = document.createElement('textarea');
        textarea.value = c.content || '';
        textarea.rows = 3;
        textarea.maxLength = 1000;
        textarea.className = 'edit-comment-textarea';
        const saveBtn = document.createElement('button');
        saveBtn.textContent = 'Save';
        saveBtn.className = 'save-comment-btn';
        const cancelBtn = document.createElement('button');
        cancelBtn.textContent = 'Cancel';
        cancelBtn.className = 'cancel-comment-btn';
        
        body.replaceWith(textarea);
        editBtn.style.display = 'none';
        deleteBtn.style.display = 'none';
        el.insertBefore(saveBtn, reacts);
        el.insertBefore(cancelBtn, reacts);

        // Save handler
        saveBtn.addEventListener('click', async () => {
          const newContent = textarea.value.trim();
          if (!newContent) {
            window.alert('Comment cannot be empty.');
            return;
          }
          saveBtn.disabled = true;
          try {
            const resp = await fetch(`${API_BASE}/comments/edit/${c.id}`, {
              method: 'PUT',
              credentials: 'include',
              headers: await getAuthHeaders(),
              body: JSON.stringify({ content: newContent }),
            });
            if (!resp.ok) throw new Error('Could not edit comment.');
            loadPost(); // Reload post to show updated comment
          } catch (err) {
            window.alert('Error: ' + err.message);
          } finally {
            saveBtn.disabled = false;
          }
        });

        // Cancel handler
        cancelBtn.addEventListener('click', () => {
          textarea.replaceWith(body);
          saveBtn.remove();
          cancelBtn.remove();
          editBtn.style.display = '';
          deleteBtn.style.display = '';
        });
      });

      // Attach Delete handler
      deleteBtn.addEventListener('click', async () => {
        if (!window.confirm('Are you sure you want to delete this comment?')) return;
        deleteBtn.disabled = true;
        try {
          const resp = await fetch(`${API_BASE}/comments/delete/${c.id}`, {
            method: 'DELETE',
            credentials: 'include',
            headers: await getAuthHeaders(),
          });
          if (!resp.ok) throw new Error('Could not delete comment.');
          loadPost(); // Reload post to show deleted comment
        } catch (err) {
          window.alert('Error: ' + err.message);
        } finally {
          deleteBtn.disabled = false;
        }
      });

      reacts.appendChild(editBtn);
      reacts.appendChild(deleteBtn);
    }
    // --- END COMMENT EDIT/DELETE CONTROLS ---

    el.append(header, body, reacts);
    return el;
  }
  
  // --- REACTION HANDLER (Unchanged) ---
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
      window.alert("Error: " + err.message);
    } finally {
      likeBtn.disabled = false;
      dislikeBtn.disabled = false;
    }
  }
  // --- END REACTION HANDLER ---

  // Initial load sequence
  loadPost();
}