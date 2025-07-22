const categoriesURL = 'http://localhost:8080/forum/api/categories';
const dropdownToggle = document.querySelector('.category-dropdown-toggle');
const dropdownContent = document.getElementById('category-tabs');
const forumContainer = document.getElementById('forumContainer');
const postTemplate = document.getElementById('post-template');

if (dropdownToggle) {
  dropdownToggle.addEventListener('click', () => {
    dropdownContent.classList.toggle('open');
    const arrow = dropdownToggle.querySelector('.dropdown-arrow');
    if (arrow) {
      arrow.style.transform = dropdownContent.classList.contains('open') ? 'rotate(180deg)' : '';
    }
  });
}

async function loadCategories() {
  try {
    const resp = await fetch(categoriesURL, { credentials: 'include' });
    if (!resp.ok) throw new Error('failed to load categories');
    const categories = await resp.json();
    renderCategories(categories);
  } catch (err) {
    console.error('Error fetching categories:', err);
    renderCategories([]);
  }
}

function renderCategories(categories) {
  dropdownContent.innerHTML = '';

  if (!categories || categories.length === 0) {
    const li = document.createElement('li');
    li.textContent = 'No categories available';
    li.className = 'no-categories';
    dropdownContent.appendChild(li);
    return;
  }

  categories.forEach(cat => {
    const li = document.createElement('li');
    const link = document.createElement('a');
    link.textContent = cat.name;
    link.href = `/user/category?id=${encodeURIComponent(cat.id)}`;
    link.className = 'category-item';
    li.appendChild(link);
    dropdownContent.appendChild(li);
  });
}

const urlParams = new URLSearchParams(window.location.search);
const categoryId = urlParams.get('id');

async function loadCategoryPosts(id) {
  try {
    const [categoryResp, feedResp] = await Promise.all([
      fetch(`http://localhost:8080/forum/api/category?id=${id}`, {
        credentials: 'include'
      }),
      fetch(`http://localhost:8080/forum/api/feed`, {
        credentials: 'include'
      })
    ]);

    if (!categoryResp.ok) {
      const errorData = await categoryResp.json();
      const code = errorData.code || categoryResp.status;
      const message = encodeURIComponent(errorData.message || errorData.error || "Unknown error");
      window.location.href = `/user/error?code=${code}&msg=${message}`;
      return;
    }

    const category = await categoryResp.json();
    const feedData = await feedResp.json();

    renderCategoryPosts(category, feedData.categories || []);
  } catch (err) {
    const fallback = encodeURIComponent("Network error or backend unreachable");
    window.location.href = `/user/error?msg=${fallback}`;
  }
}


function renderCategoryPosts(category, feedCategories) {
  forumContainer.innerHTML = '';

  const title = document.createElement('h2');
  title.textContent = category.name || `Category ${category.id}`;
  forumContainer.appendChild(title);

  if (!category.posts || category.posts.length === 0) {
    const noPosts = document.createElement('p');
    noPosts.textContent = 'No posts in this category yet.';
    noPosts.className = 'no-posts';
    forumContainer.appendChild(noPosts);
    return;
  }

  // Flatten feed posts with reactions
  const feedPosts = mergePostsFromCategories(feedCategories);

  category.posts.forEach(post => {
    // Find updated version of post in feed
    const updated = feedPosts.find(p => p.id === post.id);
    const likes = updated?.reactions?.filter(r => r.reaction_type === 1).length || 0;
    const dislikes = updated?.reactions?.filter(r => r.reaction_type === 2).length || 0;

    const postNode = postTemplate.content.cloneNode(true);
    const postEl = postNode.querySelector('.post');

    if (post.thumbnail_url) {
      const img = document.createElement('img');
      img.src = post.thumbnail_url;
      img.alt = 'Post thumbnail';
      img.className = 'post-thumb';
      postEl.insertBefore(img, postEl.firstChild);
    }

    postNode.querySelector('.post-header').textContent = post.username || 'Anonymous';
    postNode.querySelector('.post-title').textContent = post.title;
    postNode.querySelector('.post-content').textContent = post.content;
    postNode.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();
    postNode.querySelector('.like-count').textContent = likes;
    postNode.querySelector('.dislike-count').textContent = dislikes;

    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentContainer = document.createElement('span');
    commentContainer.className = 'comment-count';
    commentContainer.innerHTML = `💬 ${commentCount}`;
    postNode
      .querySelector('.like-count')
      .parentNode.appendChild(commentContainer);

    const postWrapper = document.createElement('a');
    postWrapper.href = `/user/post?id=${encodeURIComponent(post.id)}`;
    postWrapper.className = 'post-link';
    postWrapper.appendChild(postNode);

    forumContainer.appendChild(postWrapper);
  });
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
          categories: [{ id: categoryId, name: categoryName }]
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


window.addEventListener('DOMContentLoaded', () => {
  loadCategories();
  if (categoryId) {
    loadCategoryPosts(categoryId);
  } else {
    const noIdMessage = document.createElement('p');
    noIdMessage.textContent = 'No category ID provided in the URL.';
    noIdMessage.style.color = 'var(--text-muted)';
    forumContainer.appendChild(noIdMessage);
  }
});
