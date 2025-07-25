const likedContainer = document.getElementById('liked-posts-container');
const dislikedContainer = document.getElementById('disliked-posts-container');
const postTemplate = document.getElementById('post-template');

window.addEventListener('DOMContentLoaded', () => {
  loadReactions();
});

async function loadReactions() {
  likedContainer.textContent = 'Loading...';
  dislikedContainer.textContent = 'Loading...';

  const [liked, disliked] = await Promise.all([
    fetchPosts('liked'),
    fetchPosts('disliked')
  ]);

  renderPosts(liked, likedContainer, 'You have not liked any posts yet.');
  renderPosts(disliked, dislikedContainer, 'You have not disliked any posts yet.');
}

async function fetchPosts(type) {
  try {
    const resp = await fetch(`http://localhost:8080/forum/api/user/${type}`, { credentials: 'include' });
    if (!resp.ok) throw new Error('Failed to load posts');
    return await resp.json();
  } catch (err) {
    return [];
  }
}

function renderPosts(posts, container, emptyMsg) {
  container.innerHTML = '';
  if (!Array.isArray(posts) || posts.length === 0) {
    container.textContent = emptyMsg;
    return;
  }
  const fragment = document.createDocumentFragment();
  posts.forEach(post => {
    const node = postTemplate.content.cloneNode(true);
    const postEl = node.querySelector('.post');
    if (post.thumbnail_url) {
      const img = document.createElement('img');
      img.src = post.thumbnail_url;
      img.alt = 'Post thumbnail';
      img.className = 'post-thumb';
      postEl.insertBefore(img, postEl.firstChild);
    }
    if (post.title === "" && post.content === "") {
      node.querySelector('.post-title').textContent = 'This post was deleted';
      node.querySelector('.post-content').textContent = '';
      node.querySelector('.post-time').textContent = '';
  } else {
      node.querySelector('.post-header').textContent = post.username || 'Anonymous';
      node.querySelector('.post-title').textContent = post.title;
      node.querySelector('.post-content').textContent = post.content;
      node.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();
  }
    const reactionsArray = Array.isArray(post.reactions) ? post.reactions : [];
    node.querySelector('.like-count').textContent = reactionsArray.filter(r => r.reaction_type === 1).length;
    node.querySelector('.dislike-count').textContent = reactionsArray.filter(r => r.reaction_type === 2).length;
    // Wrap post in clickable link
    const wrapper = document.createElement('a');
    wrapper.href = `/user/post?id=${post.id}`;
    wrapper.className = 'post-link';
    wrapper.setAttribute('aria-label', `View post titled "${post.title}" by ${post.username}`);
    wrapper.appendChild(node);
    fragment.appendChild(wrapper);
  });
  container.appendChild(fragment);
} 