const params = new URLSearchParams(window.location.search);
const postId = params.get('id');

async function loadPost() {
  if (!postId) {
    document.getElementById('postContainer').textContent = 'Post ID missing.';
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
      document.getElementById('postContainer').textContent = 'Post not found.';
      return;
    }

    renderSinglePost(post);
  } catch (err) {
    console.error(err);
    document.getElementById('postContainer').textContent = 'Error loading post.';
  }
}

function renderSinglePost(post) {
  const container = document.getElementById('postContainer');
  container.innerHTML = '';

  const title = document.createElement('h1');
  title.className = 'post-title';
  title.textContent = post.title || 'Untitled';

  const meta = document.createElement('div');
  meta.className = 'post-meta';
  meta.textContent = `By ${post.username || post.user_id || 'Unknown'} on ${new Date(post.created_at).toLocaleString()}`;

  let imageEl = null;
  if (post.image_url) {
    imageEl = document.createElement('img');
    imageEl.src = post.image_url;
    imageEl.className = 'post-image';
  }

  const content = document.createElement('div');
  content.className = 'post-content';
  content.textContent = post.content || '';

  const reactions = document.createElement('div');
  reactions.className = 'post-reactions';
  const likes = post.reactions?.filter(r => r.reaction_type === 1).length || 0;
  const dislikes = post.reactions?.filter(r => r.reaction_type === 2).length || 0;
  reactions.innerHTML = `
    <button disabled>▲ ${likes}</button>
    <button disabled>▼ ${dislikes}</button>
  `;

  const commentCount =
    post.comment_count || (post.comments ? post.comments.length : 0);
  const commentCounter = document.createElement('span');
  commentCounter.className = 'comment-count';
  commentCounter.textContent = `💬 ${commentCount}`;
  reactions.appendChild(commentCounter);

  const categoryEl = document.createElement('div');
  categoryEl.className = 'post-categories';
  categoryEl.innerHTML = `<span class="posted-on-text">posted on the </span>`;
  post.categories?.forEach((cat, idx) => {
    const a = document.createElement('a');
    a.href = `/guest/category?id=${encodeURIComponent(cat.id)}`;
    a.textContent = cat.name;
    a.className = 'post-category-link';
    categoryEl.appendChild(a);
    if (idx < post.categories.length - 1) {
      categoryEl.appendChild(document.createTextNode(', '));
    }
  });

  // Comments
  const commentSection = document.createElement('div');
  commentSection.className = 'comments-section';
  commentSection.style.marginTop = '2rem';

  const commentHeader = document.createElement('h3');
  commentHeader.textContent = 'Comments';
  // Applying color from CSS variable directly via style property or by adding a class if defined
  // For consistency, let's assume h3 in post.css (or guest.css) will define its color.
  // If not, you'd add: commentHeader.style.color = 'var(--text-primary)';
  commentSection.appendChild(commentHeader);

  if (post.comments?.length > 0) {
    post.comments.forEach(comment => {
      const commentEl = document.createElement('div');
      commentEl.className = 'comment';
      // Removed hardcoded border-top, padding-top, margin-top to let CSS handle it
      // commentEl.style.borderTop = '1px solid #ccc';
      // commentEl.style.paddingTop = '0.5rem';
      // commentEl.style.marginTop = '0.5rem';

      const commentUser = document.createElement('strong');
      commentUser.textContent = comment.username || comment.user_id || 'Anonymous';
      // Color handled by .comment strong in post.css (var(--color-primary)) - NO JS CHANGE NEEDED

      const commentTime = document.createElement('time');
      commentTime.textContent = ` (${new Date(comment.created_at).toLocaleString()})`;
      // Color handled by .comment time in post.css (var(--text-muted)) - NO JS CHANGE NEEDED
      // commentTime.style.fontSize = '0.85em'; // Handled by CSS
      // commentTime.style.color = '#666'; // Handled by CSS (var(--text-muted))

      const commentContent = document.createElement('div');
      commentContent.textContent = comment.content || '';
      // Color handled by .comment p (if you add a p tag) or by .comment itself in post.css (var(--text-secondary))
      // commentContent.style.margin = '0.25rem 0'; // Handled by CSS

      const commentReactions = document.createElement('div');
      commentReactions.className = 'comment-reactions';
      // commentReactions.style.marginTop = '0.25rem'; // Handled by CSS

      const likeCount = comment.reactions?.filter(r => r.reaction_type === 1).length || 0;
      const dislikeCount = comment.reactions?.filter(r => r.reaction_type === 2).length || 0;

      // Buttons for comment reactions will be styled by .post-reactions button in post.css
      commentReactions.innerHTML = `
        <button disabled>▲ ${likeCount}</button>
        <button disabled>▼ ${dislikeCount}</button>
      `;

      commentEl.appendChild(commentUser);
      commentEl.appendChild(commentTime);
      commentEl.appendChild(commentContent);
      commentEl.appendChild(commentReactions);

      commentSection.appendChild(commentEl);
    });
  } else {
    const noComments = document.createElement('p');
    noComments.textContent = 'No comments yet.';
    // Ensure this text is also light. Use a CSS class for consistency or inline style.
    noComments.style.color = 'var(--text-muted)'; // Using var for consistency
    commentSection.appendChild(noComments);
  }

  container.appendChild(title);
  container.appendChild(meta);
  if (imageEl) container.appendChild(imageEl);
  container.appendChild(content);
  container.appendChild(reactions);
  container.appendChild(categoryEl);
  container.appendChild(commentSection);
}

function mergePostsFromCategories(categories) {
  const postsMap = new Map();

  categories.forEach(category => {
    const categoryId = category.id || category.ID;
    const categoryName = category.name || category.Name;

    category.posts.forEach(post => {
      const postId = post.id || post.ID;
      if (!postsMap.has(postId)) {
        postsMap.set(postId, {
          ...post,
          categories: [{ id: categoryId, name: categoryName }],
        });
      } else {
        const existingPost = postsMap.get(postId);
        if (!existingPost.categories.some(c => c.id === categoryId)) {
          existingPost.categories.push({ id: categoryId, name: categoryName });
        }
      }
    });
  });

  return Array.from(postsMap.values());
}

window.addEventListener('DOMContentLoaded', loadPost);