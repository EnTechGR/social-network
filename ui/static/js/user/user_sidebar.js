// /static/js/user_sidebar.js
const dropdownToggle = document.querySelector('.category-dropdown-toggle');
const dropdownContent = document.getElementById('category-tabs');

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
    const resp = await fetch('http://localhost:8080/forum/api/categories', {
      credentials: 'include',
    });
    if (!resp.ok) throw new Error('Failed to load categories');
    const categories = await resp.json();
    renderCategories(categories);
  } catch (err) {
    console.error('Failed to fetch categories:', err);
    renderCategories([]);
  }
}

function renderCategories(categories) {
  dropdownContent.innerHTML = '';

  if (!categories || categories.length === 0) {
    const li = document.createElement('li');
    li.textContent = 'No categories';
    li.className = 'no-categories';
    dropdownContent.appendChild(li);
    return;
  }

  categories.forEach(cat => {
    const li = document.createElement('li');
    const a = document.createElement('a');
    a.href = `/user/category?id=${encodeURIComponent(cat.id)}`;
    a.textContent = cat.name;
    a.className = 'category-item';
    li.appendChild(a);
    dropdownContent.appendChild(li);
  });
}

window.addEventListener('DOMContentLoaded', loadCategories);
