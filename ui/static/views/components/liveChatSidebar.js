// /static/components/liveChatSidebar.js
import { renderChat } from "../renderChat.js";

const API_BASE = "http://localhost:8080";
const WS_URL = "ws://localhost:8080/ws";

export function initLiveChatSidebar() {
  const conversationsList = document.getElementById("conversationsList");
  const usersList = document.getElementById("usersList");
  const currentUserInfo = document.getElementById("currentUserInfo");
  const main = document.getElementById("mainContent");

  // If layout hasn’t rendered yet or DOM IDs changed, bail safely
  if (!conversationsList || !usersList || !currentUserInfo || !main) {
    console.warn("[chat] Sidebar elements not found, skipping init");
    return;
  }

  let currentUser = null;
  let csrfToken = null;
  let ws = null;
  const onlineUsers = new Set(); // store user IDs as strings

  // ========= Utility =========
  function formatTime(date) {
    const now = new Date();
    const diff = now - date;

    if (diff < 60000) return "Just now";
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  }

  function throttle(func, delay) {
    let lastCall = 0;
    return (...args) => {
      const now = Date.now();
      if (now - lastCall >= delay) {
        lastCall = now;
        func(...args);
      }
    };
  }

  // ========= Session & init =========
  async function checkSession() {
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
      csrfToken = data.csrf_token;
      currentUserInfo.textContent = `Chat: ${currentUser.username}`;

      await initializeChatSidebar();
    } catch (err) {
      console.error("[chat] Session check failed:", err);
      currentUserInfo.textContent = "Chat unavailable";
    }
  }

  async function initializeChatSidebar() {
    await Promise.all([loadConversations(), loadAllUsers()]);
    connectWebSocket();
  }

  // ========= Conversations =========
  async function loadConversations() {
    try {
      const resp = await fetch(`${API_BASE}/forum/api/messages/conversations`, {
        credentials: "include",
      });
      if (!resp.ok) return;
      const data = await resp.json();
      renderConversations(data.conversations || []);
    } catch (err) {
      console.error("[chat] Failed to load conversations:", err);
    }
  }

  function renderConversations(convs) {
    conversationsList.innerHTML = "";

    if (!convs || convs.length === 0) {
      conversationsList.innerHTML =
        '<div class="chat-empty-small">No conversations yet</div>';
      return;
    }

    convs
      .sort(
        (a, b) => new Date(b.last_message_time) - new Date(a.last_message_time)
      )
      .forEach((conv) => {
        const item = createConversationElement(conv);
        conversationsList.appendChild(item);
      });
  }

  function createConversationElement(conv) {
    const item = document.createElement("div");
    item.className = "conversation-item";
    item.dataset.userId = String(conv.user_id);

    // Initial online state for sidebar dot
    const isOnline = conv.is_online || onlineUsers.has(String(conv.user_id));

    item.innerHTML = `
    <div class="conversation-top">
      <span class="online-indicator ${isOnline ? "online" : ""}"></span>
      <span class="conversation-username">${conv.username}</span>
    </div>
    <div class="last-message">${conv.last_message || "No messages yet"}</div>
    <div class="timestamp">${formatTime(new Date(conv.last_message_time))}</div>
  `;

    // When you click, re-check online state and pass it into renderChat
    item.addEventListener("click", () => {
      const isOnlineNow =
        conv.is_online || onlineUsers.has(String(conv.user_id));

      renderChat(main, {
        userId: conv.user_id,
        username: conv.username,
        isOnline: isOnlineNow,
      });
    });

    return item;
  }

  function updateConversationLastMessage(userId, message, timestamp) {
    const item = conversationsList.querySelector(
      `.conversation-item[data-user-id="${userId}"]`
    );

    if (!item) {
      // Conversation may be new → reload list
      loadConversations();
      return;
    }

    const lastMsg = item.querySelector(".last-message");
    const ts = item.querySelector(".timestamp");
    if (lastMsg) lastMsg.textContent = message;
    if (ts) ts.textContent = formatTime(timestamp);

    // move to top
    conversationsList.insertBefore(item, conversationsList.firstChild);
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
      renderUsersList(data.users || []);
    } catch (err) {
      console.error("[chat] Failed to load users:", err);
    }
  }

  function renderUsersList(users) {
    usersList.innerHTML = "";

    if (users.length === 0) {
      usersList.innerHTML =
        '<div class="chat-empty-small">No users available</div>';
      return;
    }

    const online = users
      .filter((u) => onlineUsers.has(String(u.id)))
      .sort((a, b) => a.username.localeCompare(b.username));
    const offline = users
      .filter((u) => !onlineUsers.has(String(u.id)))
      .sort((a, b) => a.username.localeCompare(b.username));

    [...online, ...offline].forEach((user) => {
      const item = createUserListItem(user);
      usersList.appendChild(item);
    });
  }

  function createUserListItem(user) {
    const isOnline = onlineUsers.has(String(user.id));
    const item = document.createElement("div");
    item.className = "user-item";
    item.dataset.userId = String(user.id);

    item.innerHTML = `
    <div class="online-indicator ${isOnline ? "online" : ""}"></div>
    <div class="user-details">
      <div class="user-name">${user.username}</div>
      <div class="user-email">${user.email}</div>
    </div>
  `;

    item.addEventListener("click", () => {
      const isOnlineNow = onlineUsers.has(String(user.id));

      renderChat(main, {
        userId: user.id,
        username: user.username,
        isOnline: isOnlineNow,
      });
    });

    return item;
  }

  function refreshOnlineIndicators() {
    // conversations
    conversationsList.querySelectorAll(".conversation-item").forEach((item) => {
      const uid = item.dataset.userId;
      const ind = item.querySelector(".online-indicator");
      if (!ind) return;
      if (onlineUsers.has(uid)) ind.classList.add("online");
      else ind.classList.remove("online");
    });

    // users list
    usersList.querySelectorAll(".user-item").forEach((item) => {
      const uid = item.dataset.userId;
      const ind = item.querySelector(".online-indicator");
      if (!ind) return;
      if (onlineUsers.has(uid)) ind.classList.add("online");
      else ind.classList.remove("online");
    });
  }

  // ========= WebSocket =========
  function connectWebSocket() {
    ws = new WebSocket(WS_URL);

    ws.onopen = () => {
      console.log("[chat] WebSocket connected");
    };

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      handleWebSocketMessage(msg);
    };

    ws.onerror = (err) => {
      console.error("[chat] WebSocket error:", err);
    };

    ws.onclose = () => {
      console.log("[chat] WebSocket closed, retrying...");
      setTimeout(connectWebSocket, 3000);
    };
  }

  function handleWebSocketMessage(message) {
    switch (message.type) {
      case "chat":
        handleIncomingMessage(message.data);
        break;
      case "online_status":
        handleOnlineStatus(message.data);
        break;
      case "online_users_list":
        handleOnlineUsersList(message.data);
        break;
      default:
        break;
    }
  }

  function handleIncomingMessage(data) {
    // update conversation preview
    updateConversationLastMessage(
      data.sender_id,
      data.content,
      new Date(data.created_at)
    );

    if (window.receiveChatMessage) {
      window.receiveChatMessage(data);
    }
  }

  function handleOnlineStatus(data) {
    if (data.is_online) onlineUsers.add(String(data.user_id));
    else onlineUsers.delete(String(data.user_id));

    refreshOnlineIndicators();

    // also ping open chat thread (if any)
    if (window.updateChatPartnerStatus) {
      window.updateChatPartnerStatus(data.user_id, data.is_online);
    }
  }

  function handleOnlineUsersList(users) {
    onlineUsers.clear();
    users.forEach((u) => onlineUsers.add(String(u.user_id)));
    refreshOnlineIndicators();
  }

  // kick off
  checkSession();
}
