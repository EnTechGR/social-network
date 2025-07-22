const categoriesURL = 'http://localhost:8080/forum/api/categories';
const dropdownToggle = document.querySelector('.category-dropdown-toggle');
const dropdownContent = document.getElementById('category-tabs');
const forumContainer = document.getElementById('forumContainer');
const postTemplate = document.getElementById('post-template');

// Toggle dropdown visibility
if (dropdownToggle) {
  dropdownToggle.addEventListener('click', () => {
    dropdownContent.classList.toggle('open');
    const arrow = dropdownToggle.querySelector('.dropdown-arrow');
    if (arrow) {
      arrow.style.transform = dropdownContent.classList.contains('open') ? 'rotate(180deg)' : '';
    }
  });
}

// Load and render category list in sidebar
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
    li.className = 'no-categories'; // Use CSS class for styling
    dropdownContent.appendChild(li);
    return;
  }

  categories.forEach(cat => {
    const li = document.createElement('li');
    const link = document.createElement('a');
    link.textContent = cat.name;
    link.href = `/guest/category?id=${encodeURIComponent(cat.id)}`;
    link.className = 'category-item'; // Use CSS class for styling
    li.appendChild(link);
    dropdownContent.appendChild(li);
  });
}

// Get current category ID from query string
const urlParams = new URLSearchParams(window.location.search);
const categoryId = urlParams.get('id');

// Load and render the category's posts
async function loadCategoryPosts(id) {
  try {
    const resp = await fetch(`http://localhost:8080/forum/api/category?id=${id}`, {
      credentials: 'include'
    });

    if (!resp.ok) {
      const errorData = await resp.json();
      const code = errorData.code || resp.status;
      const message = encodeURIComponent(errorData.message || errorData.error || "Unknown error");
      window.location.href = `/guest/error?code=${code}&msg=${message}`;
      return;
    }

    const category = await resp.json();
    renderCategoryPosts(category);
  } catch (err) {
    const fallback = encodeURIComponent("Network error or backend unreachable");
    window.location.href = `/guest/error?msg=${fallback}`;
  }
}

function renderCategoryPosts(category) {
  forumContainer.innerHTML = '';

  const title = document.createElement('h2');
  title.textContent = category.name || `Category ${category.id}`;
  // No need for a class here, as forum-content h2 targets it directly.
  forumContainer.appendChild(title);

  if (!category.posts || category.posts.length === 0) {
    const noPosts = document.createElement('p');
    noPosts.textContent = 'No posts in this category yet.';
    noPosts.className = 'no-posts'; // Add class for styling
    forumContainer.appendChild(noPosts);
    return;
  }

  category.posts.forEach(post => {
    const postNode = postTemplate.content.cloneNode(true);
    const postEl = postNode.querySelector('.post');

    if (post.thumbnail_url) {
      const img = document.createElement('img');
      img.src = post.thumbnail_url;
      img.alt = 'Post thumbnail';
      img.className = 'post-thumb';
      postEl.insertBefore(img, postEl.firstChild);
    }

    // Set text content for template elements
    // The classes here (post-header, post-title, etc.) are crucial for CSS to apply
    postNode.querySelector('.post-header').textContent = post.username || 'Anonymous';
    postNode.querySelector('.post-title').textContent = post.title;
    postNode.querySelector('.post-content').textContent = post.content;
    postNode.querySelector('.post-time').textContent = new Date(post.created_at).toLocaleString();

    postNode.querySelector('.like-count').textContent = post.likes || 0;
    postNode.querySelector('.dislike-count').textContent = post.dislikes || 0;

    const commentCount =
      post.comment_count || (post.comments ? post.comments.length : 0);
    const commentContainer = document.createElement('span');
    commentContainer.className = 'comment-count';
    commentContainer.innerHTML = `💬 ${commentCount}`;
    postNode
      .querySelector('.like-count')
      .parentNode.appendChild(commentContainer);

    // Wrap post in clickable anchor, applying the 'post-link' class
    const postWrapper = document.createElement('a');
    postWrapper.href = `/guest/post?id=${encodeURIComponent(post.id)}`;
    postWrapper.className = 'post-link'; // This class is key for the card styling and hover
    postWrapper.appendChild(postNode);

    forumContainer.appendChild(postWrapper);
  });
}

// Initialize both dropdown and category content
window.addEventListener('DOMContentLoaded', () => {
  loadCategories();

  if (categoryId) {
    loadCategoryPosts(categoryId);
  } else {
    // If no category ID, display a message that also uses light text.
    const noIdMessage = document.createElement('p');
    noIdMessage.textContent = 'No category ID provided in the URL.';
    noIdMessage.style.color = 'var(--text-muted)'; // Ensure this text is light
    forumContainer.appendChild(noIdMessage);
  }
});