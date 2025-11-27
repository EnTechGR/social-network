// /static/views/renderProfile.js

import { getCSRF, ensureSessionChecked, getUser } from "../session.js";

const API_BASE = "http://localhost:8080/forum/api";

export async function renderProfile(target) {
  // Ensure session is checked
  await ensureSessionChecked();
  const currentUser = getUser();

  if (!currentUser) {
    target.innerHTML = `
      <div class="profile-page fade-in">
        <p class="error-message">You must be logged in to view your profile.</p>
      </div>
    `;
    return;
  }

  // Show loading state
  target.innerHTML = `
    <div class="profile-page fade-in">
      <div class="loader">Loading profile...</div>
    </div>
  `;

  try {
    // Fetch full profile data from the API
    const response = await fetch(`${API_BASE}/user/profile`, {
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Failed to load profile");
    }

    const userData = await response.json();

    // Render the profile page
    target.innerHTML = `
      <div class="profile-page fade-in">
        <div class="profile-header">
          <div class="profile-avatar">
            ${userData.username.charAt(0).toUpperCase()}
          </div>
          <h1>My Profile</h1>
        </div>

        <div class="profile-details">
          <div class="profile-section">
            <h2>Account Information</h2>
            <div class="profile-info-grid">
              <div class="profile-info-item">
                <label>Username</label>
                <p>${escapeHtml(userData.username)}</p>
              </div>
              <div class="profile-info-item">
                <label>Email</label>
                <p>${escapeHtml(userData.email)}</p>
              </div>
              <div class="profile-info-item">
                <label>First Name</label>
                <p>${escapeHtml(userData.first_name || "N/A")}</p>
              </div>
              <div class="profile-info-item">
                <label>Last Name</label>
                <p>${escapeHtml(userData.last_name || "N/A")}</p>
              </div>
              <div class="profile-info-item">
                <label>Age</label>
                <p>${userData.age || "N/A"}</p>
              </div>
              <div class="profile-info-item">
                <label>Gender</label>
                <p>${formatGender(userData.gender)}</p>
              </div>
            </div>
          </div>

          <div class="profile-section">
            <h2>Account Details</h2>
            <div class="profile-info-grid">
              <div class="profile-info-item">
                <label>User ID</label>
                <p>#${userData.id}</p>
              </div>
              <div class="profile-info-item">
                <label>Member Since</label>
                <p>${formatDate(userData.created_at)}</p>
              </div>
            </div>
          </div>

          <div class="profile-section">
            <h2>Account Management</h2>
            <div class="profile-actions">
              <button id="logoutAllBtn" class="btn-danger">
                Logout from All Devices
              </button>
            </div>
          </div>
        </div>
      </div>
    `;

    // Add event listener for logout all button
    const logoutAllBtn = document.getElementById("logoutAllBtn");
    if (logoutAllBtn) {
      logoutAllBtn.addEventListener("click", handleLogoutAll);
    }
  } catch (error) {
    console.error("Error loading profile:", error);
    target.innerHTML = `
      <div class="profile-page fade-in">
        <div class="error-container">
          <h2>Error Loading Profile</h2>
          <p class="error-message">Unable to load your profile. Please try again later.</p>
        </div>
      </div>
    `;
  }
}

// Helper function to format gender
function formatGender(gender) {
  if (!gender) return "N/A";
  const genderMap = {
    male: "Male",
    female: "Female",
    other: "Other",
    prefer_not_to_say: "Prefer not to say",
  };
  return genderMap[gender.toLowerCase()] || gender;
}

// Helper function to format date
function formatDate(dateString) {
  if (!dateString) return "N/A";
  const date = new Date(dateString);
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

// Helper function to escape HTML
function escapeHtml(text) {
  if (!text) return "";
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

// Handle logout from all devices
async function handleLogoutAll() {
  if (
    !confirm(
      "Are you sure you want to logout from all devices? This will end all your active sessions."
    )
  ) {
    return;
  }

  const token = getCSRF();
  if (!token) {
    alert("Security token missing. Please refresh and try again.");
    return;
  }

  try {
    const response = await fetch(`${API_BASE}/session/logout-all`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": token,
      },
      credentials: "include",
    });

    if (!response.ok) {
      throw new Error("Failed to logout from all devices");
    }

    // Redirect to login page after successful logout
    alert("Successfully logged out from all devices.");
    window.location.href = "/login";
  } catch (error) {
    console.error("Error logging out from all devices:", error);
    alert("Failed to logout from all devices. Please try again.");
  }
}