import { navigateTo } from "../../router.js";

const API_BASE = "http://localhost:8080/forum/api";

export function initCategoryDropdown({
  toggleSelector,
  dropdownId,
  basePath = "/user",
}) {
  const dropdownToggle = document.querySelector(toggleSelector);
  const dropdownContent = document.getElementById(dropdownId);

  // Toggle dropdown visibility
  if (dropdownToggle) {
    dropdownToggle.addEventListener("click", () => {
      dropdownContent.classList.toggle("open");
      const arrow = dropdownToggle.querySelector(".dropdown-arrow");
      if (arrow) {
        arrow.style.transform = dropdownContent.classList.contains("open")
          ? "rotate(180deg)"
          : "";
      }
    });
  }

  async function loadCategories() {
    try {
      const resp = await fetch(`${API_BASE}/categories`, {
        credentials: "include",
      });
      if (!resp.ok) throw new Error("Failed to load categories");
      const categories = await resp.json();
      renderCategories(categories);
    } catch (err) {
      console.error("Error fetching categories:", err);
      renderCategories([]);
    }
  }

  function renderCategories(categories) {
    dropdownContent.innerHTML = "";

    if (!categories || categories.length === 0) {
      const li = document.createElement("li");
      li.textContent = "No categories available";
      li.className = "no-categories";
      dropdownContent.appendChild(li);
      return;
    }

    categories.forEach((cat) => {
      const li = document.createElement("li");
      const link = document.createElement("a");
      link.textContent = cat.name;
      link.href = `${basePath}/category/${encodeURIComponent(cat.id)}`;
      link.className = "category-item";
      link.addEventListener("click", (e) => {
        e.preventDefault();
        navigateTo(`${basePath}/category/${cat.id}`);
      });
      li.appendChild(link);
      dropdownContent.appendChild(li);
    });
  }

  return { loadCategories };
}
