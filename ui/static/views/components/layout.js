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
          <button id="profileBtn" class="profile-btn">🦄 Profile</button>
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

    <!-- ✅ Permission prompt for desktop notifications (Top Banner) -->
    <div id="notificationPermissionPrompt" class="notification-permission-prompt hidden">
      <div class="permission-prompt-content">
        <div class="permission-prompt-title">🔔 Enable Notifications</div>
        <div class="permission-prompt-text">
          Get notified when someone likes or comments on your posts, even when you're not actively using BookTalk.
        </div>
      </div>
      <div class="permission-prompt-actions">
        <button class="permission-prompt-btn deny" id="permissionDeny">Not Now</button>
        <button class="permission-prompt-btn allow" id="permissionAllow">Enable</button>
      </div>
    </div>
  `;

  // ✅ Get element references FIRST
  const createPostBtn = document.getElementById("createPostBtn");
  const notificationBtn = document.getElementById("notificationsBtn");
  const notificationBadge = document.getElementById("notificationBadge");
  const permissionPrompt = document.getElementById("notificationPermissionPrompt");
  const permissionAllow = document.getElementById("permissionAllow");
  const permissionDeny = document.getElementById("permissionDeny");

  // ✅ State variables
  let hasRequestedPermission = false;
  let notificationAudio = null;

  const catDropdown = initCategoryDropdown({
    dropdownId: "category-list",
    basePath: "/user",
  });
  catDropdown.loadCategories();

  // ✅ Preload notification sound
  preloadNotificationSound();

  // ✅ Check permission status and show prompt if needed
  checkNotificationPermissionStatus();

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

  // ✅ Permission prompt buttons
  permissionAllow.addEventListener("click", async () => {
    await requestNotificationPermission();
    hidePermissionPrompt();
  });

  permissionDeny.addEventListener("click", () => {
    hidePermissionPrompt();
    // Remember user declined (use localStorage)
    localStorage.setItem("notificationPermissionDenied", Date.now());
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

  // ✅ Preload notification sound
  function preloadNotificationSound() {
    try {
      // ✅ CHANGE THIS PATH to your audio file
      // Example paths:
      // - '/static/sounds/notification.mp3'
      // - '/static/audio/notification.ogg'
      // - '/static/notification.wav'
      const audioPath = '/static/sounds/notification.mp3';
      
      notificationAudio = new Audio(audioPath);
      notificationAudio.volume = 0.3; // 30% volume
      
      // Preload the audio file
      notificationAudio.load();
      
      console.log("[Notifications] Audio preloaded from:", audioPath);
    } catch (err) {
      console.error("[Notifications] Failed to preload audio:", err);
    }
  }

  // ✅ Check notification permission status
  function checkNotificationPermissionStatus() {
    if (!("Notification" in window)) {
      console.log("[Notifications] Desktop notifications not supported");
      return;
    }

    // Check if user already denied (and it's been less than 7 days)
    const deniedTime = localStorage.getItem("notificationPermissionDenied");
    if (deniedTime) {
      const daysSinceDenied = (Date.now() - parseInt(deniedTime)) / (1000 * 60 * 60 * 24);
      if (daysSinceDenied < 7) {
        console.log("[Notifications] User declined recently, not showing prompt");
        return;
      }
    }

    // Show prompt if permission is default (not granted or denied)
    if (Notification.permission === "default" && !hasRequestedPermission) {
      // Show our custom prompt after a short delay (better UX)
      setTimeout(() => {
        showPermissionPrompt();
      }, 3000); // Wait 3 seconds after page load
    }
  }

  // ✅ Show permission prompt
  function showPermissionPrompt() {
    permissionPrompt.classList.remove("hidden");
    permissionPrompt.classList.add("show");
  }

  // ✅ Hide permission prompt
  function hidePermissionPrompt() {
    permissionPrompt.classList.remove("show");
    setTimeout(() => {
      permissionPrompt.classList.add("hidden");
    }, 300);
  }

  // ✅ Request notification permission (called on user action)
  async function requestNotificationPermission() {
    if (!("Notification" in window)) {
      console.log("[Notifications] Desktop notifications not supported");
      return false;
    }

    if (Notification.permission === "granted") {
      console.log("[Notifications] Permission already granted");
      return true;
    }

    try {
      hasRequestedPermission = true;
      const permission = await Notification.requestPermission();
      console.log("[Notifications] Permission:", permission);
      
      if (permission === "granted") {
        // Show a test notification
        showTestNotification();
        return true;
      }
      
      return false;
    } catch (err) {
      console.error("[Notifications] Permission request failed:", err);
      return false;
    }
  }

  // ✅ Show test notification
  function showTestNotification() {
    try {
      const testNotif = new Notification("BookTalk Notifications Enabled! 🎉", {
        body: "You'll now receive notifications for likes, comments, and more.",
        icon: "/favicon.ico",
        tag: "welcome-notification"
      });

      setTimeout(() => testNotif.close(), 5000);
    } catch (err) {
      console.error("[Notifications] Failed to show test notification:", err);
    }
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

      // ✅ Play notification sound
      playNotificationSound();

      // ✅ Show desktop notification
      showDesktopNotification(notification);

      // ✅ Show toast notification
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

  // ✅ Play notification sound (using custom audio file)
  function playNotificationSound() {
    if (!notificationAudio) {
      console.warn("[Notifications] Audio not loaded");
      return;
    }

    try {
      // Reset audio to beginning
      notificationAudio.currentTime = 0;
      
      // Play the sound
      const playPromise = notificationAudio.play();
      
      if (playPromise !== undefined) {
        playPromise
          .then(() => {
            console.log("[Notifications] Sound played successfully");
          })
          .catch(err => {
            // Autoplay was prevented
            console.log("[Notifications] Sound blocked by browser:", err.message);
            // Fallback to Web Audio API if audio file is blocked
            playFallbackSound();
          });
      }
    } catch (err) {
      console.error("[Notifications] Failed to play sound:", err);
      playFallbackSound();
    }
  }

  // ✅ Fallback sound using Web Audio API (if audio file fails)
  function playFallbackSound() {
    try {
      const audioContext = new (window.AudioContext || window.webkitAudioContext)();
      const oscillator = audioContext.createOscillator();
      const gainNode = audioContext.createGain();
      
      oscillator.connect(gainNode);
      gainNode.connect(audioContext.destination);
      
      // Pleasant "ding" sound
      oscillator.frequency.setValueAtTime(800, audioContext.currentTime);
      oscillator.frequency.exponentialRampToValueAtTime(600, audioContext.currentTime + 0.1);
      
      gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);
      
      oscillator.start(audioContext.currentTime);
      oscillator.stop(audioContext.currentTime + 0.5);
      
      console.log("[Notifications] Fallback sound played");
    } catch (err) {
      console.error("[Notifications] Fallback sound failed:", err);
    }
  }

  // ✅ Show desktop notification
  function showDesktopNotification(notification) {
    // Check if notifications are supported and permitted
    if (!("Notification" in window)) {
      return;
    }

    if (Notification.permission !== "granted") {
      return;
    }

    // Don't show desktop notification if window is focused
    if (document.hasFocus()) {
      console.log("[Notifications] Window has focus, skipping desktop notification");
      return;
    }

    try {
      const message = formatNotificationMessage(notification);
      const title = "BookTalk";
      
      const options = {
        body: message,
        icon: "/favicon.ico",
        badge: "/favicon.ico",
        tag: notification.id, // Prevents duplicate notifications
        requireInteraction: false,
        silent: true, // Don't play sound (we already did)
        vibrate: [200, 100, 200],
        data: {
          url: getNotificationUrl(notification)
        }
      };

      const desktopNotif = new Notification(title, options);

      // Handle notification click
      desktopNotif.onclick = (event) => {
        event.preventDefault();
        window.focus();
        
        const url = getNotificationUrl(notification);
        if (url) {
          navigateTo(url);
        }
        
        desktopNotif.close();
      };

      // Auto-close after 5 seconds
      setTimeout(() => {
        desktopNotif.close();
      }, 5000);

      console.log("[Notifications] Desktop notification shown");
    } catch (err) {
      console.error("[Notifications] Failed to show desktop notification:", err);
    }
  }

  // ✅ Get URL for notification based on type
  function getNotificationUrl(notification) {
    if (notification.post_id) {
      return `/user/post/${notification.post_id}`;
    }
    return "/user/notifications";
  }

  // ✅ Show toast notification (in-app popup)
  function showNotificationToast(notification) {
    const toast = document.createElement("div");
    toast.className = "notification-toast";
    
    const icon = getNotificationIcon(notification.type);
    const message = formatNotificationMessage(notification);
    
    toast.innerHTML = `
      <span class="toast-icon">${icon}</span>
      <span class="toast-message">${message}</span>
    `;

    document.body.appendChild(toast);

    // Animate in
    setTimeout(() => toast.classList.add("show"), 10);

    // Make it clickable
    toast.addEventListener("click", () => {
      const url = getNotificationUrl(notification);
      if (url) {
        navigateTo(url);
      }
      toast.classList.remove("show");
      setTimeout(() => toast.remove(), 300);
    });

    // Remove after 4 seconds
    setTimeout(() => {
      toast.classList.remove("show");
      setTimeout(() => toast.remove(), 300);
    }, 4000);
  }

  // ✅ Get icon for notification type
  function getNotificationIcon(type) {
    switch (type) {
      case "like":
        return "👍";
      case "dislike":
        return "👎";
      case "love":
        return "❤️";
      case "comment":
        return "💬";
      case "edit_comment":
        return "✏️";
      case "delete_comment":
        return "🗑️";
      default:
        return "🔔";
    }
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