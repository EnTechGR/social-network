
// Helper to get query param
function getQueryParam(name) {
    const url = new URL(window.location.href);
    return url.searchParams.get(name);
}

const postId = getQueryParam('id');
const postContainer = document.getElementById('postContainer');

if (!postId) {
    postContainer.textContent = 'No post ID provided.';
} else {
    loadPost();
}

// Helper to determine deleted post display state
function getPostDisplayState(post) {
    let isDeleted = false;
    let displayTitle = post.title;
    let displayContent = post.content;
    if ((post.title === "") && (post.content === "")) {
        displayTitle = 'This post was deleted';
        displayContent = null;
        isDeleted = true;
    }
    return { isDeleted, displayTitle, displayContent };
}

async function loadPost() {
    if (!postId) {
      postContainer.textContent = 'Post ID missing.';
      return;
    }
  
    try {
      const resp = await fetch('http://localhost:8080/forum/api/feed', {
        credentials: 'include',
      });
  
      if (!resp.ok) throw new Error('Failed to load post');
  
      const data = await resp.json();
      const posts = mergePostsFromCategories(data.categories || []);
      const post = posts.find(p => p.id === postId);
  
      if (!post) {
        postContainer.textContent = 'Post not found.';
        return;
      }
  
      renderSinglePostWithEdit(post);
    } catch (err) {
      console.error(err);
      postContainer.textContent = 'Error loading post.';
    }
  }

// Render the post with interactive like/dislike buttons & comments
function renderSinglePostWithEdit(post) {
    postContainer.innerHTML = '';

    // --- Deleted post check ---
    const { isDeleted, displayTitle, displayContent } = getPostDisplayState(post);
    
    // --- Add Delete control ---
    const controls = document.createElement('div');
    controls.className = 'post-controls';

    const deleteBtn = document.createElement('button');
    deleteBtn.textContent = 'Delete';
    deleteBtn.className = 'delete-btn';
    deleteBtn.onclick = deletePost;
    // --- End controls ---


    const title = document.createElement('h1');
    title.className = isDeleted ? 'deleted-title' : 'post-title';
    title.textContent = displayTitle;

    // Add edit button for title
    const titleEditBtn = document.createElement('button');
    titleEditBtn.textContent = 'Edit Title';
    titleEditBtn.className = 'edit-btn';
    titleEditBtn.onclick = () => {
        showEditTitle(post.title);
    };
    if (isDeleted) titleEditBtn.style.display = 'none';

    const meta = document.createElement('div');
    meta.className = 'post-meta';
    let metaDate, isEdited = false;
    if (post.updated_at && post.updated_at !== post.created_at) {
        metaDate = new Date(post.updated_at).toLocaleString();
        isEdited = true;
    } else {
        metaDate = new Date(post.created_at).toLocaleString();
    }

    let label = "";
    if (isDeleted) {
        label = " (Deleted)";
    } else if (isEdited) {
        label = " (Edited)";
    }

    meta.textContent = `By ${post.username || post.user_id || 'Unknown'} on ${metaDate}${label}`;

    const content = document.createElement('div');
    content.className = isDeleted ? 'deleted-content' : 'post-content';
    content.textContent = displayContent;

    // Add edit button for content
    const contentEditBtn = document.createElement('button');
    contentEditBtn.textContent = 'Edit Content';
    contentEditBtn.className = 'edit-btn';
    contentEditBtn.onclick = () => {
        showEditContent(post.content);
    };
    if (isDeleted) contentEditBtn.style.display = 'none';

          let imageEl = null;
     if (post.image_url) {
       imageEl = document.createElement('img');
       imageEl.src = post.image_url;
       imageEl.className = 'post-image';
     }

     // Add edit button for image
     const imageEditBtn = document.createElement('button');
     imageEditBtn.textContent = 'Edit Image';
     imageEditBtn.className = 'edit-btn';
     imageEditBtn.onclick = () => {
         showEditImage();
     };
     if (isDeleted || !post.image_url) imageEditBtn.style.display = 'none';

    // Wrap post content in a card
    const postContentCard = document.createElement('div');
    postContentCard.className = 'post-content-card';
    if (content) postContentCard.appendChild(content);

    // Reactions container with interactive buttons
    const reactions = document.createElement('div');
    reactions.className = 'post-reactions';
  
    // Count likes & dislikes
    const reactionsArray = Array.isArray(post.reactions) ? post.reactions : [];
    let likes = reactionsArray.filter(r => r.reaction_type === 1).length || 0;
    let dislikes = reactionsArray.filter(r => r.reaction_type === 2).length || 0;
  
    const likeBtn = document.createElement('button');
    likeBtn.textContent = `▲ ${likes}`;
    likeBtn.className = 'like-btn';
    likeBtn.title = 'Like';
    if (isDeleted) likeBtn.disabled = true;

    const dislikeBtn = document.createElement('button');
    dislikeBtn.textContent = `▼ ${dislikes}`;
    dislikeBtn.className = 'dislike-btn';
    dislikeBtn.title = 'Dislike';
    if (isDeleted) dislikeBtn.disabled = true;
  
    reactions.appendChild(likeBtn);
    reactions.appendChild(dislikeBtn);
  
    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentCounter = document.createElement('span');
    commentCounter.className = 'comment-count';
    commentCounter.textContent = `💬 ${commentCount}`;
    reactions.appendChild(commentCounter);
  
    // Reaction button click handlers
    likeBtn.addEventListener('click', () => handleReaction(post.id, 'post', 1, likeBtn, dislikeBtn));
    dislikeBtn.addEventListener('click', () => handleReaction(post.id, 'post', 2, likeBtn, dislikeBtn));
  
    // Categories
    const categoryEl = document.createElement('div');
    categoryEl.className = 'post-categories';
    categoryEl.innerHTML = `<span class="Posted-on-text">Posted on the </span>`;
    post.categories?.forEach((cat, idx) => {
      const a = document.createElement('a');
      a.href = `/user/category?id=${encodeURIComponent(cat.id)}`;
      a.textContent = cat.name;
      a.className = 'post-category-link';
      categoryEl.appendChild(a);
      if (idx < post.categories.length - 1) {
        categoryEl.appendChild(document.createTextNode(', '));
      }
    });
  
    // Comments Section
    const commentSection = document.createElement('div');
    commentSection.className = 'comments-section';
  
    const commentHeader = document.createElement('h3');
    commentHeader.textContent = 'Comments';
    commentSection.appendChild(commentHeader);
  
    // Inline Comment Form (no modal, always visible)
    const commentFormContainer = document.createElement('div');
    commentFormContainer.className = 'comment-form-container';
  
    const commentForm = document.createElement('form');
    commentForm.className = 'comment-form';
    commentForm.autocomplete = 'off';
  
    const commentTextarea = document.createElement('textarea');
    commentTextarea.className = 'comment-textarea';
    commentTextarea.placeholder = "Write your comment...";
    commentTextarea.required = true;
    commentTextarea.rows = 3;
    commentTextarea.maxLength = 1000;
  
    const submitCommentBtn = document.createElement('button');
    submitCommentBtn.type = 'submit';
    submitCommentBtn.className = 'submit-comment-btn';
    submitCommentBtn.textContent = 'Submit Comment';
  
    // Error message element for comment form
    const errorMsg = document.createElement('div');
    errorMsg.className = 'comment-error-msg';
  
    // Character count element
    const charCount = document.createElement('div');
    charCount.className = 'comment-char-count';
    charCount.textContent = '0 / 1000';
  
    // Update character count on input
    commentTextarea.addEventListener('input', () => {
      charCount.textContent = `${commentTextarea.value.length} / 1000`;
      if (commentTextarea.value.length > 1000) {
        commentTextarea.value = commentTextarea.value.slice(0, 1000);
      }
      errorMsg.classList.remove('visible');
    });
  
    // Insert elements in the form
    commentForm.appendChild(errorMsg);
    commentForm.appendChild(commentTextarea);
    commentForm.appendChild(charCount);
    commentForm.appendChild(submitCommentBtn);
  
    // Submit comment handler (with validation)
    commentForm.addEventListener('submit', async (e) => {
      e.preventDefault();
      const content = commentTextarea.value.trim();
      if (!content) {
        errorMsg.textContent = 'Comment cannot be empty.';
        errorMsg.classList.add('visible');
        return;
      }
      if (content.length > 1000) {
        errorMsg.textContent = 'Comment cannot exceed 1000 characters.';
        errorMsg.classList.add('visible');
        return;
      }
      errorMsg.classList.remove('visible');
      if (!csrfTokenFromResponse) {
        csrfTokenFromResponse = await loadCSRFTokenFromSession();
        if (!csrfTokenFromResponse) {
          alert('Session expired or not authenticated. Please log in again.');
          return;
        }
      }
      submitCommentBtn.disabled = true;
      submitCommentBtn.textContent = 'Submitting...';
      try {
        const resp = await fetch('http://localhost:8080/forum/api/comments/create', {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrfTokenFromResponse,
          },
          body: JSON.stringify({
            post_id: post.id,
            content,
          }),
        });
        if (!resp.ok) {
          const errData = await resp.json().catch(() => ({}));
          errorMsg.textContent = 'Error: ' + (errData.message || 'Could not submit comment.');
          errorMsg.classList.add('visible');
          return;
        }
        commentTextarea.value = '';
        errorMsg.classList.remove('visible');
        await loadPost();
      } catch (err) {
        console.error('Failed to submit comment:', err);
        errorMsg.textContent = 'Failed to submit comment. Try again later.';
        errorMsg.classList.add('visible');
      } finally {
        submitCommentBtn.disabled = false;
        submitCommentBtn.textContent = 'Submit Comment';
      }
    });
  
    commentFormContainer.appendChild(commentForm);
  
    // Comments list
    if (post.comments?.length > 0) {
      post.comments.forEach(comment => {
        commentSection.appendChild(createCommentElement(comment, isDeleted));
      });
    } else {
      const noComments = document.createElement('p');
      noComments.textContent = 'No comments yet.';
      noComments.className = 'no-comments';
      commentSection.appendChild(noComments);
    }
  
    // Create a boxed post container
    const postBox = document.createElement('div');
    postBox.className = 'post';

    postBox.appendChild(title);
    if (!isDeleted) postBox.appendChild(titleEditBtn);
         postBox.appendChild(meta);
      if (imageEl) postBox.appendChild(imageEl);
     if (imageEl && !isDeleted) postBox.appendChild(imageEditBtn);
    if (!isDeleted) {
        postBox.appendChild(postContentCard);
        postBox.appendChild(contentEditBtn);
        postBox.appendChild(commentFormContainer);
    }
    postBox.appendChild(reactions);
    postBox.appendChild(categoryEl);
    if (!isDeleted) {
        postBox.appendChild(deleteBtn);
    }
    postBox.appendChild(commentSection);
  
    // Add everything to the DOM
    postContainer.appendChild(postBox);
  }
  
  // Helper: create comment element with reactions
  function createCommentElement(comment, isPostDeleted) {

    // Match guest style: compact, simple, but keep interactive buttons
    const commentEl = document.createElement('div');
    commentEl.className = 'comment';
  
    const commentUser = document.createElement('strong');
    commentUser.textContent = comment.username || comment.user_id || 'Anonymous';
  
    const commentTime = document.createElement('time');
    if (comment.content === "") {
      commentTime.textContent = ` (${new Date(comment.updated_at).toLocaleString()})`;
    } else {
      commentTime.textContent = ` (${new Date(comment.created_at).toLocaleString()})`;
    }
  
    const commentContent = document.createElement('div');
    if (comment.content === "") {
      commentContent.textContent = 'This comment was deleted';
    } else {
      commentContent.textContent = comment.content || '';
    }
    commentContent.className = 'comment-content';

    // Add (Edited) label to the date if the comment was edited
    let isEdited = false;
    if (comment.updated_at && comment.updated_at !== comment.created_at) {
      isEdited = true;
    }
    if (commentTime) {
      if (comment.content === "") {
        commentTime.textContent = ` (${new Date(comment.updated_at).toLocaleString()})`;
      } else if (isEdited) {
        commentTime.textContent = ` (${new Date(comment.updated_at).toLocaleString()}) (Edited)`;
      } else {
        commentTime.textContent = ` (${new Date(comment.created_at).toLocaleString()})`;
      }
    }

    // Reactions: visually match guest (inline, compact, no extra box)
    const commentReactions = document.createElement('div');
    commentReactions.className = 'comment-reactions';
  
    const reactionsArray = Array.isArray(comment.reactions) ? comment.reactions : [];
    const likeCount = reactionsArray.filter(r => r.reaction_type === 1).length || 0;
    const dislikeCount = reactionsArray.filter(r => r.reaction_type === 2).length || 0;
  
    const likeBtn = document.createElement('button');
    likeBtn.textContent = `▲ ${likeCount}`;
    likeBtn.className = 'like-btn';
    likeBtn.title = 'Like';
    if (isPostDeleted) likeBtn.disabled = true;
  
    const dislikeBtn = document.createElement('button');
    dislikeBtn.textContent = `▼ ${dislikeCount}`;
    dislikeBtn.className = 'dislike-btn';
    dislikeBtn.title = 'Dislike';
    if (isPostDeleted) dislikeBtn.disabled = true;

    // Attach handlers for comment reactions (keep interactive)
    likeBtn.addEventListener('click', () => handleReaction(comment.id, 'comment', 1, likeBtn, dislikeBtn));
    dislikeBtn.addEventListener('click', () => handleReaction(comment.id, 'comment', 2, likeBtn, dislikeBtn));
  
    commentReactions.appendChild(likeBtn);
    commentReactions.appendChild(dislikeBtn);

    // --- Edit/Delete for own comments ---
    if (currentUserId && comment.user_id === currentUserId && !isPostDeleted) {
      // Edit button
      const editBtn = document.createElement('button');
      editBtn.textContent = 'Edit';
      editBtn.className = 'edit-comment-btn';
      editBtn.addEventListener('click', () => {
        // Replace content with textarea and save/cancel buttons
        const textarea = document.createElement('textarea');
        textarea.value = comment.content || '';
        textarea.rows = 3;
        textarea.maxLength = 1000;
        textarea.className = 'edit-comment-textarea';
        const saveBtn = document.createElement('button');
        saveBtn.textContent = 'Save';
        saveBtn.className = 'save-comment-btn';
        const cancelBtn = document.createElement('button');
        cancelBtn.textContent = 'Cancel';
        cancelBtn.className = 'cancel-comment-btn';
        // Replace content
        commentContent.replaceWith(textarea);
        editBtn.style.display = 'none';
        deleteBtn.style.display = 'none';
        commentEl.insertBefore(saveBtn, commentReactions);
        commentEl.insertBefore(cancelBtn, commentReactions);
        // Save handler
        saveBtn.addEventListener('click', async () => {
          const newContent = textarea.value.trim();
          if (!newContent) {
            alert('Comment cannot be empty.');
            return;
          }
          if (!csrfTokenFromResponse) {
            csrfTokenFromResponse = await loadCSRFTokenFromSession();
            if (!csrfTokenFromResponse) {
              alert('Session expired. Please log in again.');
              return;
            }
          }
          saveBtn.disabled = true;
          try {
            const resp = await fetch(`http://localhost:8080/forum/api/comments/edit/${comment.id}`, {
              method: 'PUT',
              credentials: 'include',
              headers: await getAuthHeaders(),
              body: JSON.stringify({ content: newContent }),
            });
            if (!resp.ok) {
              const errData = await resp.json().catch(() => ({}));
              alert('Error: ' + (errData.message || 'Could not edit comment.'));
              return;
            }
            loadPost();
          } finally {
            saveBtn.disabled = false;
          }
        });
        // Cancel handler
        cancelBtn.addEventListener('click', () => {
          textarea.replaceWith(commentContent);
          saveBtn.remove();
          cancelBtn.remove();
          editBtn.style.display = '';
          deleteBtn.style.display = '';
        });
      });
      // Delete button
      const deleteBtn = document.createElement('button');
      deleteBtn.textContent = 'Delete';
      deleteBtn.className = 'delete-comment-btn';
      deleteBtn.addEventListener('click', async () => {
        if (!confirm('Are you sure you want to delete this comment?')) return;
        if (!csrfTokenFromResponse) {
          csrfTokenFromResponse = await loadCSRFTokenFromSession();
          if (!csrfTokenFromResponse) {
            alert('Session expired. Please log in again.');
            return;
          }
        }
        deleteBtn.disabled = true;
        try {
          const resp = await fetch(`http://localhost:8080/forum/api/comments/delete/${comment.id}`, {
            method: 'DELETE',
            credentials: 'include',
            headers: await getAuthHeaders(),
          });
          if (!resp.ok) {
            const errData = await resp.json().catch(() => ({}));
            alert('Error: ' + (errData.message || 'Could not delete comment.'));
            return;
          }
          loadPost();
        } finally {
          deleteBtn.disabled = false;
        }
      });
      if (comment.content === "") {
        editBtn.style.display = 'none';
        deleteBtn.style.display = 'none';
      }
      commentReactions.appendChild(editBtn);
      commentReactions.appendChild(deleteBtn);
    }
    // --- End edit/delete ---

    // Layout: username, time, content, reactions (all compact)
    commentEl.appendChild(commentUser);
    commentEl.appendChild(commentTime);
    commentEl.appendChild(commentContent);
    commentEl.appendChild(commentReactions);
  
    return commentEl;
  }
// --- End: Copied rendering logic ---

function showEditTitle(currentTitle) {
    // Find the title element and replace it with edit form
    const titleElement = document.querySelector('.post-title, .deleted-title');
    if (titleElement) {
        titleElement.innerHTML = `<input id="titleInput" value="${currentTitle ?? ''}" style="width:60%;-webkit-text-fill-color:black"/> <button id="saveTitleBtn" style="-webkit-text-fill-color:black">Save</button> <button id="cancelTitleBtn" style="-webkit-text-fill-color:black">Cancel</button>`;
        document.getElementById('saveTitleBtn').onclick = saveTitle;
        document.getElementById('cancelTitleBtn').onclick = loadPost;
    }
}

async function saveTitle() {
    const newTitle = document.getElementById('titleInput').value;
    await fetch(`http://localhost:8080/forum/api/posts/edit-title/${postId}`, {
        method: 'PUT',
        headers: await getAuthHeaders(),
        body: JSON.stringify({ title: newTitle }),
        credentials: 'include'
    });
    loadPost();
}

function showEditContent(currentContent) {
    // Find the content element and replace it with edit form
    const contentElement = document.querySelector('.post-content, .deleted-content');
    if (contentElement) {
        contentElement.innerHTML = `<textarea id="contentInput" style="width:90%">${currentContent ?? ''}</textarea><br/><button id="saveContentBtn">Save</button> <button id="cancelContentBtn">Cancel</button>`;
        document.getElementById('saveContentBtn').onclick = saveContent;
        document.getElementById('cancelContentBtn').onclick = loadPost;
    }
}

async function saveContent() {
    const newContent = document.getElementById('contentInput').value;
    await fetch(`http://localhost:8080/forum/api/posts/edit-content/${postId}`, {
        method: 'PUT',
        headers: await getAuthHeaders(),
        body: JSON.stringify({ content: newContent }),
        credentials: 'include'
    });
    loadPost();
}

function showEditImage() {
    // Find the image element and add upload interface right after it
    const imageElement = document.querySelector('.post-image');
    if (imageElement) {
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
                
                const resp = await fetch('http://localhost:8080/forum/api/images/upload', {
                    method: 'POST',
                    headers: await getAuthHeaders(true),
                    body: formData,
                    credentials: 'include'
                });
                
                if (!resp.ok) {
                    const errorData = await resp.json().catch(() => ({}));
                    throw new Error(errorData.message || 'Upload failed');
                }
                
                // Remove upload interface and reload the post to show the new image
                uploadContainer.remove();
                loadPost();
            } catch (error) {
                console.error('Image upload failed:', error);
                imageError.textContent = `Upload failed: ${error.message}`;
                uploadImageBtn.disabled = false;
                uploadImageBtn.textContent = "Upload Image";
            }
        });
    }
}

async function deletePost() {
    if (!confirm('Are you sure you want to delete this post?')) return;
    try {
        // Delete the post (soft-delete)
        await fetch(`http://localhost:8080/forum/api/posts/delete/${postId}`, {
            method: 'DELETE',
            headers: await getAuthHeaders(),
            credentials: 'include'
        });
        // Delete the images for the post
        await fetch(`http://localhost:8080/forum/api/images/delete/${postId}`, {
            method: 'DELETE',
            headers: await getAuthHeaders(),
            credentials: 'include'
        });
        // Refresh the page after successful deletion
        window.location.reload();
    } catch (error) {
        console.error('Error deleting post:', error);
        alert('Failed to delete post. Please try again.');
    }
}

// Helper to get auth headers (including CSRF)
async function getAuthHeaders(skipContentType) {
    const headers = {};
    if (!skipContentType) headers['Content-Type'] = 'application/json';
    if (csrfTokenFromResponse) {
        headers['X-CSRF-Token'] = csrfTokenFromResponse;
    }
    return headers;
}
let csrfTokenFromResponse = null;
const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';

async function loadCSRFTokenFromSession() {
  try {
    const resp = await fetch(sessionVerifyURL, {
      credentials: 'include',
    });
    if (!resp.ok) throw new Error('Session not valid');
    const data = await resp.json();
    return data.csrf_token || data.CSRFToken;
  } catch (err) {
    console.warn("Failed to load CSRF token from session:", err);
    return null;
  }
}

let currentUserId = null;

// Fetch current user info at page load
async function fetchCurrentUserId() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/session/verify', {
      credentials: 'include',
    });
    if (!resp.ok) return null;
    const data = await resp.json();
    return data.user?.id || (data.user && data.user.ID) || null;
  } catch (err) {
    return null;
  }
}

// Helper: merge posts from categories (same as your original)
function mergePostsFromCategories(categories) {
    const postsMap = new Map();
    categories.forEach(category => {
      const categoryId = category.id;
      const categoryName = category.name;
  
      category.posts.forEach(post => {
        if (!postsMap.has(post.id)) {
          postsMap.set(post.id, {
            ...post,
            categories: [{ id: categoryId, name: categoryName }],
          });
        } else {
          const existing = postsMap.get(post.id);
          existing.categories.push({ id: categoryId, name: categoryName });
        }
      });
    });
    return Array.from(postsMap.values());
}

// On page load, fetch CSRF token and current user ID, then load post
(async () => {
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
  currentUserId = await fetchCurrentUserId();
  if (!csrfTokenFromResponse) {
    alert("Session expired or not authenticated. Please log in again.");
    return;
  }
  loadPost();
})();

