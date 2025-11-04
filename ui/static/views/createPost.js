import { getCSRF, verifySession } from "../session.js";
import { navigateTo } from "../router.js";
import { renderLayout } from "../views/components/layout.js";

const API_BASE = "http://localhost:8080/forum/api";

export function renderCreatePost(app) {
  // Render within the main app layout
  renderLayout(app, {
    mainContent: `
      <section class="create-post-page fade-in">
        <h2>Create New Post</h2>

        <form id="createPostForm" class="create-post-form">
          <label>
            Title <span id="titleCount" class="count">0 / 200</span>
            <input type="text" id="titleInput" maxlength="200" required />
          </label>

          <label>
            Content <span id="bodyCount" class="count">0 / 2000</span>
            <textarea id="bodyInput" maxlength="2000" required></textarea>
          </label>

          <div class="categories">
            <h3>Select Categories</h3>
            <div id="categoryList" class="checkbox-list"></div>
          </div>

          <div class="image-upload">
            <label>Attach Image (optional)</label>
            <input type="file" id="imageInput" accept="image/jpeg,image/png,image/gif" />
            <img id="imagePreview" class="hidden" alt="Preview" />
            <p id="imageError" class="error"></p>
          </div>

          <button type="submit" class="btn-accent">Publish Post</button>
          <p id="formMessage" class="message"></p>
        </form>
      </section>
    `,
  });

  const form = document.getElementById("createPostForm");
  const titleInput = document.getElementById("titleInput");
  const bodyInput = document.getElementById("bodyInput");
  const titleCount = document.getElementById("titleCount");
  const bodyCount = document.getElementById("bodyCount");
  const categoryList = document.getElementById("categoryList");
  const imageInput = document.getElementById("imageInput");
  const imagePreview = document.getElementById("imagePreview");
  const imageError = document.getElementById("imageError");
  const message = document.getElementById("formMessage");

  // Character counters
  titleInput.addEventListener("input", () => {
    titleCount.textContent = `${titleInput.value.length} / 200`;
  });
  bodyInput.addEventListener("input", () => {
    bodyCount.textContent = `${bodyInput.value.length} / 2000`;
  });

  // Image preview + validation
  imageInput.addEventListener("change", () => {
    const file = imageInput.files[0];
    if (!file) {
      imagePreview.classList.add("hidden");
      imageError.textContent = "";
      return;
    }

    if (!["image/jpeg", "image/png", "image/gif"].includes(file.type)) {
      imageError.textContent = "Unsupported file type. Use JPG, PNG, or GIF.";
      imageInput.value = "";
      return;
    }

    if (file.size > 20 * 1024 * 1024) {
      imageError.textContent = "Image exceeds 20MB limit.";
      imageInput.value = "";
      return;
    }

    imageError.textContent = "";
    const reader = new FileReader();
    reader.onload = (e) => {
      imagePreview.src = e.target.result;
      imagePreview.classList.remove("hidden");
    };
    reader.readAsDataURL(file);
  });

  // Load categories dynamically (with toggle behavior)
  async function loadCategories() {
    try {
      const res = await fetch(`${API_BASE}/categories`, {
        credentials: "include",
      });
      if (!res.ok) throw new Error("Failed to load categories");
      const categories = await res.json();
      const container = document.createElement("div");
      container.classList.add("category-tags");

      categories.forEach((cat) => {
        const tag = document.createElement("div");
        tag.textContent = cat.name;
        tag.className = "category-tag";
        tag.dataset.id = cat.id;
        tag.addEventListener("click", () => {
          tag.classList.toggle("active");
        });
        container.appendChild(tag);
      });

      categoryList.innerHTML = "";
      categoryList.appendChild(container);
    } catch (err) {
      console.error("Error loading categories:", err);
      categoryList.innerHTML = `<p class="error">Unable to load categories.</p>`;
    }
  }

  loadCategories();

  // Handle form submission
  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    message.textContent = "";
    message.className = "message";

    const title = titleInput.value.trim();
    const content = bodyInput.value.trim();
    const categoryIds = Array.from(
      document.querySelectorAll(".category-tag.active")
    ).map((tag) => parseInt(tag.dataset.id));

    if (!title || !content || categoryIds.length === 0) {
      message.textContent =
        "Please fill out all fields and select at least one category.";
      message.classList.add("error");
      return;
    }

    // Ensure CSRF token
    await verifySession();
    const csrfToken = getCSRF();
    if (!csrfToken) {
      message.textContent = "Session expired. Please log in again.";
      message.classList.add("error");
      return navigateTo("/login");
    }

    // Submit post
    try {
      const res = await fetch(`${API_BASE}/posts/create`, {
        method: "POST",
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrfToken,
        },
        body: JSON.stringify({ title, content, category_ids: categoryIds }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message || "Failed to create post");
      }

      const newPost = await res.json();

      // Upload image if provided
      if (imageInput.files.length > 0) {
        const formData = new FormData();
        formData.append("post_id", newPost.id);
        formData.append("image", imageInput.files[0]);

        const imgRes = await fetch(`${API_BASE}/images/upload`, {
          method: "POST",
          credentials: "include",
          headers: {
            "X-CSRF-Token": csrfToken,
          },
          body: formData,
        });

        if (!imgRes.ok) console.warn("Image upload failed");
      }

      message.textContent = "Post created successfully!";
      message.classList.add("success");

      setTimeout(() => navigateTo("/user/feed"), 800);
    } catch (err) {
      console.error("Error submitting post:", err);
      message.textContent = err.message || "Error creating post.";
      message.classList.add("error");
    }
  });
}
