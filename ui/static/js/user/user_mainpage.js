const categoriesURL = 'http://localhost:8080/forum/api/categories';

const dropdownToggle = document.querySelector('.category-dropdown-toggle');
const dropdownContent = document.getElementById('category-tabs');
const forumContainer = document.getElementById('forumContainer');

// Dropdown open/close
if (dropdownToggle) {
  dropdownToggle.addEventListener('click', () => {
    dropdownContent.classList.toggle('open');
    const arrow = dropdownToggle.querySelector('.dropdown-arrow');
    if (arrow) {
      arrow.style.transform = dropdownContent.classList.contains('open') ? 'rotate(180deg)' : '';
    }
  });
}

// Fetch and display categories
async function loadCategories() {
  try {
    const resp = await fetch(categoriesURL, { credentials: 'include' });
    if (!resp.ok) throw new Error('Failed to load categories');
    const categories = await resp.json();
    renderCategories(categories);
  } catch (err) {
    console.error('Error loading categories:', err);
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

// Handle logout on back/forward navigation
window.addEventListener('popstate', async () => {
  try {
    await fetch('http://localhost:8080/forum/api/session/logout', { method: 'POST', credentials: 'include' });
    window.location.href = '/';
  } catch (e) {
    console.error('Logout failed', e);
  }
});

const logoutLink = document.getElementById('logout-link');

if (logoutLink) {
  logoutLink.addEventListener('click', async (e) => {
    e.preventDefault(); // prevent default anchor navigation
    try {
      const res = await fetch('http://localhost:8080/forum/api/session/logout', {
        method: 'POST',
        credentials: 'include'
      });
      if (res.ok) {
        window.location.href = '/login';
      } else {
        console.error('Logout failed with status:', res.status);
      }
    } catch (err) {
      console.error('Logout error:', err);
    }
  });
}


// Initialize on DOM ready
window.addEventListener('DOMContentLoaded', loadCategories);
