import { renderChat } from "../renderChat.js";

const API_BASE = "http://localhost:8080";
const WS_URL = "ws://localhost:8080/ws";

let ws = null;
let shouldReconnect = false;
let cleanupRegistered = false;
let isConnecting = false; // ✅ ADD: Prevent multiple simultaneous connection attempts

export function initLiveChatSidebar() {
  const main = document.getElementById("mainContent");

  if (!main) {
    console.warn("[chat] Main content element not found, skipping init");
    return;
  }

  let currentUser = null;

  const onlineUsers = new Set();
  const conversationsMeta = new Map();
  const unreadCounts = new Map();
  let lastUsers = [];
  let activeChatUserId = null;

  // ========= Utility =========
  function formatTime(date) {
    const now = new Date();
    const diff = now - date;
    if (diff < 60000) return "Just now";
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  }

  // ========= Session & init =========
  async function checkSession() {
    const currentUserInfo = document.getElementById("currentUserInfo");
    if (!currentUserInfo) return;

    try {
      const resp = await fetch(`${API_BASE}/forum/api/session/verify`, {
        credentials: "include",
      });

      if (!resp.ok) {
        currentUserInfo.textContent = "Chat: login to start conversations";
        return;
      }

      const data = await resp.json();
      currentUser = data.user;
      currentUserInfo.textContent = `${currentUser.username}`;

      shouldReconnect = true;
      await initializeChatSidebar();
    } catch (err) {
      console.error("[chat] Session check failed:", err);
      currentUserInfo.textContent = "Chat unavailable";
    }
  }

  async function initializeChatSidebar() {
    await Promise.all([loadConversations(), loadAllUsers()]);
    
    // ✅ FIX: Only connect if not already connected or connecting
    if (!ws || ws.readyState === WebSocket.CLOSED) {
      connectWebSocket();
    }

    if (!cleanupRegistered) {
      window.addEventListener("beforeunload", () => {
        shouldReconnect = false; // ✅ FIX: Prevent reconnection on page unload
        if (ws) {
          ws.close();
          ws = null;
        }
      });
      cleanupRegistered = true;
    }
  }

  // ========= Conversations metadata =========
  async function loadConversations() {
    try {
      const resp = await fetch(`${API_BASE}/forum/api/messages/conversations`, {
        credentials: "include",
      });
      if (!resp.ok) return;

      const data = await resp.json();
      const convs = data.conversations || [];

      conversationsMeta.clear();
      convs.forEach((conv) => {
        conversationsMeta.set(String(conv.user_id), conv);
      });
    } catch (err) {
      console.error("[chat] Failed to load conversations:", err);
    }
  }

  function updateConversationLastMessage(userId, message, timestamp) {
    const key = String(userId);
    const existing = conversationsMeta.get(key) || {};
    conversationsMeta.set(key, {
      ...existing,
      user_id: userId,
      username: existing.username || "",
      last_message: message,
      last_message_time: timestamp.toISOString(),
    });

    if (lastUsers.length) {
      renderUsersList(lastUsers);
    }
  }

  // ========= Users list =========
  async function loadAllUsers() {
    try {
      const resp = await fetch(
        `${API_BASE}/forum/api/messages/users-for-chat?all=true`,
        { credentials: "include" }
      );
      if (!resp.ok) return;
      const data = await resp.json();
      lastUsers = data.users || [];
      renderUsersList(lastUsers);
    } catch (err) {
      console.error("[chat] Failed to load users:", err);
    }
  }

  function sortUsersForSection(users) {
    return users.slice().sort((a, b) => {
      const convA = conversationsMeta.get(String(a.id));
      const convB = conversationsMeta.get(String(b.id));

      if (convA && convB) {
        const tA = new Date(convA.last_message_time);
        const tB = new Date(convB.last_message_time);
        if (tA > tB) return -1;
        if (tA < tB) return 1;
      }

      if (convA && !convB) return -1;
      if (!convA && convB) return 1;

      return a.username.localeCompare(b.username);
    });
  }

  function renderUsersList(users) {
    const usersList = document.getElementById("usersList");
    if (!usersList) return;

    usersList.innerHTML = "";

    if (users.length === 0) {
      usersList.innerHTML =
        '<div class="chat-empty-small">No users available</div>';
      return;
    }

    const onlineList = users.filter((u) => onlineUsers.has(String(u.id)));
    const offlineList = users.filter((u) => !onlineUsers.has(String(u.id)));

    const sortedOnline = sortUsersForSection(onlineList);
    const sortedOffline = sortUsersForSection(offlineList);

    if (sortedOnline.length > 0) {
      const header = document.createElement("div");
      header.className = "chat-section-header";
      header.textContent = "Online";
      usersList.appendChild(header);

      sortedOnline.forEach((user) => {
        usersList.appendChild(createUserListItem(user));
      });
    }

    if (sortedOffline.length > 0) {
      const header = document.createElement("div");
      header.className = "chat-section-header";
      header.textContent = "Offline";
      usersList.appendChild(header);

      sortedOffline.forEach((user) => {
        usersList.appendChild(createUserListItem(user));
      });
    }
  }

  function createUserListItem(user) {
    const isOnline = onlineUsers.has(String(user.id));
    const unread = unreadCounts.get(String(user.id)) || 0;

    const item = document.createElement("div");
    item.className = "user-item";
    item.dataset.userId = String(user.id);

    item.innerHTML = `
      <div class="user-row-main">
        <div class="online-indicator ${isOnline ? "online" : ""}"></div>
        <div class="user-details">
          <div class="user-name">${user.username}</div>
          <div class="user-email">${user.email}</div>
        </div>
        ${
          unread > 0
            ? `<div class="unread-badge">${unread > 9 ? "9+" : unread}</div>`
            : ""
        }
      </div>
    `;

    item.addEventListener("click", () => {
      const isOnlineNow = onlineUsers.has(String(user.id));

      activeChatUserId = user.id;
      unreadCounts.delete(String(user.id));
      if (lastUsers.length) {
        renderUsersList(lastUsers);
      }

      renderChat(main, {
        userId: user.id,
        username: user.username,
        isOnline: isOnlineNow,
      });
    });

    return item;
  }

  function refreshOnlineIndicators() {
    const usersList = document.getElementById("usersList");
    if (!usersList) return;

    usersList.querySelectorAll(".user-item").forEach((item) => {
      const uid = item.dataset.userId;
      const ind = item.querySelector(".online-indicator");
      if (!ind) return;
      if (onlineUsers.has(uid)) ind.classList.add("online");
      else ind.classList.remove("online");
    });
  }

  window.clearActiveChatUser = () => {
    activeChatUserId = null;
  };

  // ========= WebSocket =========
  function connectWebSocket() {
    // ✅ FIX: Prevent multiple simultaneous connections
    if (isConnecting) {
      console.log("[chat] Already connecting, skipping...");
      return;
    }

    // ✅ FIX: Don't reconnect if already connected
    if (ws && (ws.readyState === WebSocket.CONNECTING || ws.readyState === WebSocket.OPEN)) {
      console.log("[chat] WebSocket already connected or connecting");
      return;
    }

    isConnecting = true;
    console.log("[chat] Initiating WebSocket connection...");
    
    ws = new WebSocket(WS_URL);

    ws.onopen = () => {
      isConnecting = false;
      console.log("[chat] ✅ WebSocket connected successfully");
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      handleWebSocketMessage(msg);
    };

    ws.onerror = (err) => {
      isConnecting = false;
      console.error("[chat] ❌ WebSocket error:", err);
    };

    ws.onclose = (event) => {
      isConnecting = false;
      console.log("[chat] 🔌 WebSocket closed. Code:", event.code, "Reason:", event.reason);
      
      // ✅ FIX: Only reconnect if it's an unexpected closure and we should reconnect
      // Code 1000 = normal closure, don't reconnect
      // Code 1001 = going away (e.g., tab closing), don't reconnect
      if (shouldReconnect && event.code !== 1000 && event.code !== 1001) {
        console.log("[chat] ⏳ Reconnecting in 3 seconds...");
        setTimeout(() => {
          if (shouldReconnect) { // Check again in case user logged out
            connectWebSocket();
          }
        }, 3000);
      } else {
        console.log("[chat] Not reconnecting (shouldReconnect:", shouldReconnect, "code:", event.code, ")");
      }
    };
  }

  function handleWebSocketMessage(message) {
    console.log('[WS] Received message:', message.type);
    
    switch (message.type) {
      case "chat":
        handleIncomingMessage(message.data);
        break;
      // ✅ FIX: Handle chat_image type
      case "chat_image":
        console.log('[WS] Handling chat_image message');
        handleIncomingMessage(message.data);
        break;
      case "online_status":
        handleOnlineStatus(message.data);
        break;
      case "online_users_list":
        handleOnlineUsersList(message.data);
        break;
      case "message_delete":
        if (window.handleMessageDelete) {
          window.handleMessageDelete(message.data);
        }
        break;
      default:
        console.warn('[WS] Unknown message type:', message.type);
        break;
    }
  }

  function handleIncomingMessage(data) {
    console.log('[WS] Processing incoming message. Has image?', !!data.image);
    
    const senderId = String(data.sender_id);

    // For image messages, show a preview in the conversation list
    const displayContent = data.image ? '📷 Image' : data.content;
    
    updateConversationLastMessage(
      senderId,
      displayContent,
      new Date(data.created_at)
    );

    if (!activeChatUserId || String(activeChatUserId) !== senderId) {
      const prev = unreadCounts.get(senderId) || 0;
      unreadCounts.set(senderId, prev + 1);
      if (lastUsers.length) {
        renderUsersList(lastUsers);
      }
    }

    if (window.receiveChatMessage) {
      window.receiveChatMessage(data);
    }
  }

  function handleOnlineStatus(data) {
    const id = String(data.user_id);

    if (data.is_online) onlineUsers.add(id);
    else onlineUsers.delete(id);

    refreshOnlineIndicators();

    if (lastUsers.length) {
      renderUsersList(lastUsers);
    }

    if (window.updateChatPartnerStatus) {
      window.updateChatPartnerStatus(id, data.is_online);
    }
  }

  function handleOnlineUsersList(users) {
    onlineUsers.clear();
    users.forEach((u) => {
      onlineUsers.add(String(u.user_id));
    });

    refreshOnlineIndicators();

    if (lastUsers.length) {
      renderUsersList(lastUsers);
    }
  }

  // kick off
  checkSession();
}

/**
 * Called on logout to stop presence + chat for this tab.
 */
export function disconnectLiveChatSidebar() {
  try {
    shouldReconnect = false;
    if (ws) {
      console.log("[chat] Closing WebSocket on logout");
      ws.close(1000, "User logged out"); // ✅ FIX: Use normal closure code
      ws = null;
    }
  } catch (err) {
    console.error("[chat] Error closing WebSocket:", err);
  }
}