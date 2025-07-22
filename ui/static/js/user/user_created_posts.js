const forumContainer = document.getElementById('forumContainer');
const postTemplate = document.getElementById('post-template');

window.addEventListener('DOMContentLoaded', () => {
  fetchCreatedPosts();
});

async function fetchCreatedPosts() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/user/posts', {
      credentials: 'include',
    });

    if (!resp.ok) {
      const err = await resp.json();
      throw new Error(err.message || 'Failed to load created posts');
    }

    const posts = await resp.json();
    renderCreatedPosts(posts);
  } catch (err) {
    console.error(`Error: ${err.message}`);
    forumContainer.textContent = 'You have not created any posts yet.';
  }
}

function renderCreatedPosts(posts) {
  forumContainer.innerHTML = '';

  if (!posts.length) {
    forumContainer.textContent = 'You have not created any posts yet.';
    return;
  }

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

    node.querySelector('.post-header').textContent = post.username || 'You';
    node.querySelector('.post-title').textContent = post.title;
    node.querySelector('.post-content').textContent = post.content;
    node.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();

    const likeCount = (post.reactions || []).filter(r => r.reaction_type === 1).length;
    const dislikeCount = (post.reactions || []).filter(r => r.reaction_type === 2).length;

    node.querySelector('.like-count').textContent = likeCount;
    node.querySelector('.dislike-count').textContent = dislikeCount;

    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentContainer = document.createElement('span');
    commentContainer.className = 'comment-count';
    commentContainer.innerHTML = `💬 ${commentCount}`;
    node
      .querySelector('.like-count')
      .parentNode.appendChild(commentContainer);

    const wrapper = document.createElement('a');
    wrapper.href = `/user/post?id=${post.id}`;
    wrapper.className = 'post-link';
    wrapper.appendChild(node);

    forumContainer.appendChild(wrapper);
  });
}
