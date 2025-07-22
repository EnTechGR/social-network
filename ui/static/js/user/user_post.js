const params = new URLSearchParams(window.location.search);
const postId = params.get('id');
const lastReactions = new Map(); // key = `${type}:${id}`, value = 1 (like) or 2 (dislike)


const postContainer = document.getElementById('postContainer');

// CSRF token in-memory storage
let csrfTokenFromResponse = null;

const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';

// Utility: load CSRF token by verifying session
async function loadCSRFTokenFromSession() {
  try {
    const resp = await fetch(sessionVerifyURL, {
      credentials: 'include',
    });
    if (!resp.ok) throw new Error('Session not valid');
    const data = await resp.json();
    return data.csrf_token || data.CSRFToken; // Adjust key if needed
  } catch (err) {
    console.warn("Failed to load CSRF token from session:", err);
    return null;
  }
}

// Fetch post feed and find the single post by id
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

    renderSinglePost(post);
  } catch (err) {
    console.error(err);
    postContainer.textContent = 'Error loading post.';
  }
}

// Render the post with interactive like/dislike buttons & comments
function renderSinglePost(post) {
  postContainer.innerHTML = '';

  const title = document.createElement('h1');
  title.className = 'post-title';
  title.textContent = post.title || 'Untitled';

  const meta = document.createElement('div');
  meta.className = 'post-meta';
  meta.textContent = `By ${post.username || post.user_id || 'Unknown'} on ${new Date(post.created_at).toLocaleString()}`;

  const content = document.createElement('div');
  content.className = 'post-content';
  content.textContent = post.content || '';

   let imageEl = null;
  if (post.image_url) {
    imageEl = document.createElement('img');
    imageEl.src = post.image_url;
    imageEl.className = 'post-image';
  }

  // Wrap post content in a card
  const postContentCard = document.createElement('div');
  postContentCard.className = 'post-content-card';
  postContentCard.appendChild(content);

  // Reactions container with interactive buttons
  const reactions = document.createElement('div');
  reactions.className = 'post-reactions';

  // Count likes & dislikes
  let likes = post.reactions?.filter(r => r.reaction_type === 1).length || 0;
  let dislikes = post.reactions?.filter(r => r.reaction_type === 2).length || 0;

  const likeBtn = document.createElement('button');
  likeBtn.textContent = `▲ ${likes}`;
  likeBtn.className = 'like-btn';
  likeBtn.title = 'Like';

  const dislikeBtn = document.createElement('button');
  dislikeBtn.textContent = `▼ ${dislikes}`;
  dislikeBtn.className = 'dislike-btn';
  dislikeBtn.title = 'Dislike';

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
      commentSection.appendChild(createCommentElement(comment));
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
  postBox.appendChild(meta);
   if (imageEl) postBox.appendChild(imageEl);
  postBox.appendChild(postContentCard); // instead of content
  postBox.appendChild(reactions);
  postBox.appendChild(categoryEl);
  postBox.appendChild(commentFormContainer);
  postBox.appendChild(commentSection);

  // Add everything to the DOM
  postContainer.appendChild(postBox);
}

// Helper: create comment element with reactions
function createCommentElement(comment) {
  // Match guest style: compact, simple, but keep interactive buttons
  const commentEl = document.createElement('div');
  commentEl.className = 'comment';

  const commentUser = document.createElement('strong');
  commentUser.textContent = comment.username || comment.user_id || 'Anonymous';

  const commentTime = document.createElement('time');
  commentTime.textContent = ` (${new Date(comment.created_at).toLocaleString()})`;

  const commentContent = document.createElement('div');
  commentContent.textContent = comment.content || '';

  // Reactions: visually match guest (inline, compact, no extra box)
  const commentReactions = document.createElement('div');
  commentReactions.className = 'comment-reactions';

  const likeCount = comment.reactions?.filter(r => r.reaction_type === 1).length || 0;
  const dislikeCount = comment.reactions?.filter(r => r.reaction_type === 2).length || 0;

  const likeBtn = document.createElement('button');
  likeBtn.textContent = `▲ ${likeCount}`;
  likeBtn.className = 'like-btn';
  likeBtn.title = 'Like';

  const dislikeBtn = document.createElement('button');
  dislikeBtn.textContent = `▼ ${dislikeCount}`;
  dislikeBtn.className = 'dislike-btn';
  dislikeBtn.title = 'Dislike';

  // Attach handlers for comment reactions (keep interactive)
  likeBtn.addEventListener('click', () => handleReaction(comment.id, 'comment', 1, likeBtn, dislikeBtn));
  dislikeBtn.addEventListener('click', () => handleReaction(comment.id, 'comment', 2, likeBtn, dislikeBtn));

  commentReactions.appendChild(likeBtn);
  commentReactions.appendChild(dislikeBtn);

  // Layout: username, time, content, reactions (all compact)
  commentEl.appendChild(commentUser);
  commentEl.appendChild(commentTime);
  commentEl.appendChild(commentContent);
  commentEl.appendChild(commentReactions);

  return commentEl;
}

// Handle like/dislike interaction
async function handleReaction(targetId, targetType, reactionType, likeBtn, dislikeBtn) {
  const key = `${targetType}:${targetId}`;
  const prevReaction = lastReactions.get(key);

  let finalReactionType = reactionType;

  // If clicking same reaction again, it means "remove reaction"
  const isRemoving = prevReaction === reactionType;
  if (isRemoving) {
    finalReactionType = 3; // Special type = remove
  }

  if (!csrfTokenFromResponse) {
    csrfTokenFromResponse = await loadCSRFTokenFromSession();
    if (!csrfTokenFromResponse) {
      alert('Session expired. Please log in again.');
      return;
    }
  }

  likeBtn.disabled = true;
  dislikeBtn.disabled = true;

  try {
    const resp = await fetch('http://localhost:8080/forum/api/react', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': csrfTokenFromResponse,
      },
      body: JSON.stringify({
        target_id: targetId,
        target_type: targetType,
        reaction_type: finalReactionType,
      }),
    });

    if (!resp.ok) {
      const errData = await resp.json().catch(() => ({}));
      throw new Error(errData.message || 'Failed to react');
    }

    const reactions = await resp.json();
    const likes = reactions.filter(r => r.reaction_type === 1).length;
    const dislikes = reactions.filter(r => r.reaction_type === 2).length;

    likeBtn.textContent = `▲ ${likes}`;
    dislikeBtn.textContent = `▼ ${dislikes}`;

    // Update local reaction state
    if (finalReactionType === 3) {
      lastReactions.delete(key); // removed
    } else {
      lastReactions.set(key, reactionType); // updated
    }

  } catch (err) {
    console.error(err);
    alert('Error: ' + err.message);
  } finally {
    likeBtn.disabled = false;
    dislikeBtn.disabled = false;
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

// Initial load
loadPost();
