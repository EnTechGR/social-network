// /static/components/initLayout.js
import { initCategoryDropdown } from "./categories.js";
import { navigateTo } from "../../router.js";
import { clearSession } from "../../session.js";
import {
  initLiveChatSidebar,
  disconnectLiveChatSidebar,
} from "../components/liveChatSidebar.js";

export function initLayout() {
  const app = document.getElementById("app");

  app.innerHTML = `
    <div class="layout">
      <header class="app-header">
        <h1 class="logo" id="homeBtn">BookTalk</h1>
        <div class="header-actions">
          <button id="createPostBtn" class="btn-accent">+ New Post</button>
          <button id="notificationsBtn" class="notification-btn">
            🔔
            <span id="notificationBadge" class="notification-badge hidden">0</span>
          </button>
          <button id="profileBtn" class="profile-btn">🐦 Profile</button>
          <button id="logoutBtn">Logout</button>
        </div>
      </header>

      <div class="app-body">
        <aside class="sidebar">
          <h2>Categories</h2>
          <ul id="category-list" class="category-list"></ul>

          <div class="user-menu">
            <h3>Your Space</h3>
            <ul id="userActivityLinks"> 
              <li><a href="/user/my-activity/my-posts">My Posts</a></li>
              <li><a href="/user/my-activity/my-reactions">My Reactions</a></li>
              <li><a href="/user/my-activity/my-comments">My Comments</a></li>
            </ul>
          </div>
        </aside>

        <main id="mainContent" class="main-content">
          <div class="loader">Loading...</div>
        </main>

        <aside class="right-panel">
          <div id="chatSidebar" class="chat-sidebar">
            <h3>Live Chat</h3>
            <div id="currentUserInfo" class="current-user-info">
              Loading...
            </div>

            <h4>Users</h4>
            <div id="usersList" class="chat-users-list"></div>
          </div>
        </aside>
      </div>
    </div>
  `;

  // ✅ Get element references FIRST
  const createPostBtn = document.getElementById("createPostBtn");
  const notificationBtn = document.getElementById("notificationsBtn");
  const notificationBadge = document.getElementById("notificationBadge");

  const catDropdown = initCategoryDropdown({
    dropdownId: "category-list",
    basePath: "/user",
  });
  catDropdown.loadCategories();

  // ✅ Load initial notification count
  loadNotificationCount();

  // ✅ Setup WebSocket listener for real-time notifications
  setupNotificationListener();

  // Listener for Home button
  document.getElementById("homeBtn").addEventListener("click", () => {
    navigateTo("/user/feed");
  });

  // Create Post Button
  createPostBtn.addEventListener("click", () => {
    navigateTo("/user/posts/create");
  });

  // Profile Button
  document.getElementById("profileBtn").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/user/profile");
  });

  // Notifications Button
  notificationBtn.addEventListener("click", (e) => {
    e.preventDefault();
    // ✅ Clear badge when viewing notifications
    notificationBadge.textContent = "0";
    notificationBadge.classList.add("hidden");
    navigateTo("/user/notifications");
  });

  // Logout Button
  document.getElementById("logoutBtn").addEventListener("click", async () => {
    disconnectLiveChatSidebar();
    await fetch("http://localhost:8080/forum/api/session/logout", {
      method: "POST",
      credentials: "include",
    });
    clearSession();
    navigateTo("/login");
  });

  // Intercept user activity links
  const userActivityLinks = document.getElementById("userActivityLinks");
  if (userActivityLinks) {
    userActivityLinks.addEventListener("click", (e) => {
      const link = e.target.closest("a");
      if (link) {
        e.preventDefault();
        const path = link.getAttribute("href");
        if (path) {
          navigateTo(path);
        }
      }
    });
  }

  // ✅ Load initial notification count from API
  async function loadNotificationCount() {
    try {
      const resp = await fetch("http://localhost:8080/forum/api/notifications", {
        credentials: "include",
      });

      if (!resp.ok) return;

      const data = await resp.json();
      const count = data.count || 0;

      updateNotificationBadge(count);
    } catch (err) {
      console.error("[Notifications] Failed to load count:", err);
    }
  }

  // ✅ Setup WebSocket listener for real-time notifications
  function setupNotificationListener() {
    // Register global handler for notification messages
    window.handleNotificationReceived = (notification) => {
      console.log("[Notifications] Received:", notification);

      // Get current badge count
      const currentCount = parseInt(notificationBadge.textContent) || 0;
      const newCount = currentCount + 1;

      updateNotificationBadge(newCount);

      // Optional: Show a toast notification
      showNotificationToast(notification);
    };
  }

  // ✅ Update the notification badge
  function updateNotificationBadge(count) {
    if (count > 0) {
      notificationBadge.textContent = count > 99 ? "99+" : count.toString();
      notificationBadge.classList.remove("hidden");
      
      // Add animation
      notificationBadge.classList.add("badge-pulse");
      setTimeout(() => {
        notificationBadge.classList.remove("badge-pulse");
      }, 300);
    } else {
      notificationBadge.textContent = "0";
      notificationBadge.classList.add("hidden");
    }
  }

  // ✅ Optional: Show a brief toast notification
  function showNotificationToast(notification) {
    const toast = document.createElement("div");
    toast.className = "notification-toast";
    
    const message = formatNotificationMessage(notification);
    toast.textContent = message;

    document.body.appendChild(toast);

    // Animate in
    setTimeout(() => toast.classList.add("show"), 10);

    // Remove after 4 seconds
    setTimeout(() => {
      toast.classList.remove("show");
      setTimeout(() => toast.remove(), 300);
    }, 4000);
  }

  // ✅ Format notification message for display
  function formatNotificationMessage(n) {
    const actor = n.username || "Someone";
    const onComment = n.comment_id !== undefined && n.comment_id !== null;

    switch (n.type) {
      case "like":
        return onComment
          ? `${actor} liked your comment`
          : `${actor} liked your post`;
      case "dislike":
        return onComment
          ? `${actor} disliked your comment`
          : `${actor} disliked your post`;
      case "love":
        return onComment
          ? `${actor} loved your comment`
          : `${actor} loved your post`;
      case "comment":
        return `${actor} commented on your post`;
      case "edit_comment":
        return `${actor} edited a comment on your post`;
      case "delete_comment":
        return `${actor} deleted a comment on your post`;
      default:
        return `${actor} interacted with your post`;
    }
  }

  initLiveChatSidebar();
}