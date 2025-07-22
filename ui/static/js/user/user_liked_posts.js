const forumContainer = document.getElementById("forumContainer");
const postTemplate = document.getElementById("post-template");

window.addEventListener("DOMContentLoaded", () => {
  fetchLikedPosts();
});

async function fetchLikedPosts() {
  forumContainer.textContent = 'Loading...';

  try {
    const resp = await fetch("http://localhost:8080/forum/api/user/liked", {
      credentials: "include",
    });

    if (!resp.ok) {
      const err = await resp.json();
      throw new Error(err.message || "Failed to load liked posts");
    }

    const posts = await resp.json(); // No filtering needed

    renderLikedPosts(posts);
  } catch (err) {
    console.error(`Error: ${err.message}`);
    forumContainer.textContent = "You have not liked any posts yet.";
  }
}

function renderLikedPosts(posts) {
  forumContainer.innerHTML = "";

  if (!posts.length) {
    forumContainer.textContent = "You have not liked any posts yet.";
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

    // Fill in post data
    node.querySelector('.post-header').textContent = post.username || 'Anonymous';
    node.querySelector('.post-title').textContent = post.title;
    node.querySelector('.post-content').textContent = post.content;
    node.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();

    // Count reactions
    let likeCount = 0, dislikeCount = 0;
    const reactions = Array.isArray(post.reactions) ? post.reactions : [];
    reactions.forEach(r => {
      if (r.reaction_type === 1) likeCount++;
      else if (r.reaction_type === 2) dislikeCount++;
    });

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

    // Wrap post in clickable link
    const wrapper = document.createElement('a');
    wrapper.href = `/user/post?id=${post.id}`;
    wrapper.className = 'post-link';
    wrapper.setAttribute('aria-label', `View post titled "${post.title}" by ${post.username}`);
    wrapper.appendChild(node);

    fragment.appendChild(wrapper);
  });

  forumContainer.appendChild(fragment);
}
