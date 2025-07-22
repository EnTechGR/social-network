const categoriesURL = "http://localhost:8080/forum/api/categories";
const sessionVerifyURL = "http://localhost:8080/forum/api/session/verify"; // This is good

const createBtn = document.getElementById("create-post-btn");
const modal = document.getElementById("post-modal");
const closeModalBtn = modal.querySelector(".close-btn");
const submitPostBtn = document.getElementById("submit-post");

const titleInput = document.getElementById("post-title");
const contentInput = document.getElementById("post-body");
const categoryCheckboxesContainer = document.getElementById("post-category");
const titleCount = document.getElementById("post-title-count");
const bodyCount = document.getElementById("post-body-count");
const imageInput = document.getElementById("post-image");
const addImageBtn = document.getElementById("add-image-btn");
const cancelImageBtn = document.getElementById("cancel-image-btn");
const imageStatus = document.getElementById("image-status");
const imageError = document.getElementById("image-error");
const imagePreview = document.getElementById("image-preview");

// Store CSRF token in-memory here (initially empty)
let csrfTokenFromResponse = null;

// Utility: load CSRF token by verifying session
async function loadCSRFTokenFromSession() {
  try {
    // Try to get from a specific frontend cookie if the Go server set it
    const csrfCookie = document.cookie
      .split("; ")
      .find((row) => row.startsWith("csrf_token_frontend="));
    if (csrfCookie) {
      return csrfCookie.split("=")[1];
    }

    // If not found in a specific frontend cookie, then make the API call
    const resp = await fetch(sessionVerifyURL, {
      credentials: "include",
    });
    if (!resp.ok) {
      const errorText = await resp.text();
      console.error("Session verify API failed:", resp.status, errorText);
      throw new Error("Session not valid");
    }
    const data = await resp.json();
    // Backend currently returns `csrf_token`, make sure it's consistent
    return data.csrf_token;
  } catch (err) {
    console.warn("Failed to load CSRF token from session:", err);
    return null;
  }
}

function updateTitleCount() {
  let val = titleInput.value;
  if (val.length > 200) {
    titleInput.value = val.slice(0, 200);
    val = titleInput.value;
  }
  titleCount.textContent = `${val.length} / 200`;
}

function updateBodyCount() {
  let val = contentInput.value;
  if (val.length > 2000) {
    contentInput.value = val.slice(0, 2000);
    val = contentInput.value;
  }
  bodyCount.textContent = `${val.length} / 2000`;
}

titleInput.addEventListener("input", updateTitleCount);
contentInput.addEventListener("input", updateBodyCount);

addImageBtn.addEventListener("click", () => imageInput.click());
cancelImageBtn.addEventListener("click", resetImageSelection);

function resetImageSelection() {
  imageInput.value = "";
  imageStatus.textContent = "";
  imageStatus.classList.add("hidden");
  imageStatus.classList.remove("status-valid", "status-error");
  imageError.textContent = "";
  cancelImageBtn.classList.add("hidden");
  addImageBtn.disabled = false;
  imagePreview.src = "";
  imagePreview.classList.add("hidden");
}

function validateSelectedImage() {
  imageError.textContent = "";
  const file = imageInput.files[0];
  if (!file) {
    // No image selected, this is allowed
    return true;
  }
  const allowed = ["image/jpeg", "image/png", "image/gif"];
  if (!allowed.includes(file.type)) {
    imageStatus.textContent = file.name;
    imageStatus.classList.remove("hidden", "status-valid");
    imageStatus.classList.add("status-error");
    imageError.textContent = "Unsupported image type. Only jpeg, png, gif";
    imageInput.value = "";
    imagePreview.src = "";
    imagePreview.classList.add("hidden");
    cancelImageBtn.classList.remove("hidden");
    return false;
  }
  if (file.size > 20 * 1024 * 1024) {
    imageStatus.textContent = file.name;
    imageStatus.classList.remove("hidden", "status-valid");
    imageStatus.classList.add("status-error");
    imageError.textContent = "Image exceeds 20 MB limit";
    imageInput.value = "";
    imagePreview.src = "";
    imagePreview.classList.add("hidden");
    cancelImageBtn.classList.remove("hidden");
    return false;
  }

  imageStatus.textContent = file.name;
  imageStatus.classList.remove("hidden", "status-error");
  imageStatus.classList.add("status-valid");
  cancelImageBtn.classList.remove("hidden");
  addImageBtn.disabled = true;
  const reader = new FileReader();
  reader.onload = (e) => {
    imagePreview.src = e.target.result;
    imagePreview.classList.remove("hidden");
  };
  reader.readAsDataURL(file);
  return true;
}

imageInput.addEventListener("change", () => {
  // Re-validate image and enable/disable submit button accordingly
  if (validateSelectedImage()) {
    submitPostBtn.disabled = false;
  } else {
    submitPostBtn.disabled = true;
  }
});

// Open modal and load categories
createBtn.addEventListener("click", async (e) => {
  e.preventDefault();
  modal.classList.remove("hidden");

  // Load CSRF token when opening the modal, to ensure it's fresh
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
  if (!csrfTokenFromResponse) {
    alert("Session expired or not authenticated. Please log in again.");
    modal.classList.add("hidden"); // Hide modal if no token
    return;
  }

  try {
    const resp = await fetch(categoriesURL, { credentials: "include" });
    if (!resp.ok) {
      const errorText = await resp.text();
      throw new Error(
        `Failed to fetch categories: ${resp.status} - ${errorText}`,
      );
    }
    const categories = await resp.json();

    categoryCheckboxesContainer.innerHTML = "";
    categories.forEach((cat) => {
      const label = document.createElement("label");
      label.classList.add("checkbox-item");

      const checkbox = document.createElement("input");
      checkbox.type = "checkbox";
      checkbox.value = cat.id;

      label.appendChild(checkbox);
      label.appendChild(document.createTextNode(" " + cat.name));
      categoryCheckboxesContainer.appendChild(label);
    });
  } catch (err) {
    console.error("Failed to load categories in modal:", err);
    categoryCheckboxesContainer.innerHTML =
      "<p class='error'>Failed to load categories</p>";
  }

  updateTitleCount();
  updateBodyCount();
});

// Close modal
closeModalBtn.addEventListener("click", () => {
  modal.classList.add("hidden");
  clearModalInputs();
});

window.addEventListener("click", (e) => {
  if (e.target === modal) {
    modal.classList.add("hidden");
    clearModalInputs();
  }
});

window.addEventListener("keydown", (e) => {
  if (e.key === "Escape" && !modal.classList.contains("hidden")) {
    modal.classList.add("hidden");
    clearModalInputs();
  }
});

function clearModalInputs() {
  titleInput.value = "";
  contentInput.value = "";
  updateTitleCount();
  updateBodyCount();
  categoryCheckboxesContainer
    .querySelectorAll("input[type=checkbox]")
    .forEach((cb) => (cb.checked = false));
  resetImageSelection();
}

// Submit post
submitPostBtn.addEventListener("click", async () => {
  const title = titleInput.value.trim();
  const content = contentInput.value.trim();
  const categoryIDs = Array.from(
    categoryCheckboxesContainer.querySelectorAll(
      "input[type=checkbox]:checked",
    ),
  ).map((cb) => parseInt(cb.value));

  if (!title || !content || categoryIDs.length === 0) {
    alert("Please fill out all fields and select at least one category.");
    return;
  }

  if (!validateSelectedImage()) {
    submitPostBtn.disabled = true;
    return;
  }

  // CSRF token should have been loaded when modal opened.
  // If for some reason it's still null (e.g., user opened modal, waited for session to expire, then clicked submit),
  // try to reload it, but it's better to ensure it's loaded on modal open.
  if (!csrfTokenFromResponse) {
    csrfTokenFromResponse = await loadCSRFTokenFromSession();
    if (!csrfTokenFromResponse) {
      alert("Session expired or not authenticated. Please log in again.");
      return;
    }
  }

  submitPostBtn.disabled = true;
  submitPostBtn.textContent = "Submitting...";

  try {
    const resp = await fetch("http://localhost:8080/forum/api/posts/create", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        // Ensure the header name matches what your backend middleware expects
        "X-CSRF-Token": csrfTokenFromResponse,
      },
      body: JSON.stringify({
        title,
        content,
        category_ids: categoryIDs,
      }),
    });

    if (!resp.ok) {
      const errData = await resp.json().catch(() => ({}));
      console.error("Backend error response:", errData);
      alert(
        "Error creating post: " +
          (errData.message || `Status: ${resp.status} - ${resp.statusText}`),
      );
      return;
    }

    const createdPost = await resp.json();

    if (imageInput.files.length > 0) {
      const formData = new FormData();
      formData.append("post_id", createdPost.id || createdPost.ID);
      formData.append("image", imageInput.files[0]);

      const imgResp = await fetch(
        "http://localhost:8080/forum/api/images/upload",
        {
          method: "POST",
          credentials: "include",
          headers: {
            "X-CSRF-Token": csrfTokenFromResponse,
          },
          body: formData,
        },
      );

      if (!imgResp.ok) {
        const errImg = await imgResp.json().catch(() => ({}));
        console.error("Image upload failed:", errImg);
        alert("Image upload failed");
      }
    }

    modal.classList.add("hidden");
    clearModalInputs();
    location.reload();
  } catch (err) {
    console.error("Post creation failed:", err);
    alert("Failed to create post. Try again later.");
  } finally {
    submitPostBtn.disabled = false;
    submitPostBtn.textContent = "Submit";
  }
});
