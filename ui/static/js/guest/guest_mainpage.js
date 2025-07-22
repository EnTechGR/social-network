// Guest page logic to fetch categories and handle dropdown

// API endpoint for categories
const categoriesURL = 'http://localhost:8080/forum/api/categories';

// Elements
const dropdownToggle = document.querySelector('.category-dropdown-toggle');
const dropdownContent = document.getElementById('category-tabs');

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

// Fetch categories from backend and populate list
async function loadCategories() {
  try {
    const resp = await fetch(categoriesURL, { credentials: 'include' });
    if (!resp.ok) {
      throw new Error('failed to load categories');
    }
    const categories = await resp.json();
    renderCategories(categories);
  } catch (err) {
    console.error('Error fetching categories:', err);
    renderCategories([]);
  }
}

// Render category items
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
    link.href = `/guest/category?id=${encodeURIComponent(cat.id)}`; // ✅ dynamic link
    link.className = 'category-item';

    li.appendChild(link);
    dropdownContent.appendChild(li);
  });
}


// Initialize on DOM ready
window.addEventListener('DOMContentLoaded', loadCategories);
