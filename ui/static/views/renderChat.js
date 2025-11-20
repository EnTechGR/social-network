// /static/renderChat.js

import { getCSRF, ensureSessionChecked, getUser } from "../session.js";

const API_BASE = "http://localhost:8080/forum/api";

export async function renderChat(main, { userId, username, isOnline = false }) {
  // Make sure session & CSRF are loaded
  await ensureSessionChecked();
  const currentUser = getUser();

  if (!currentUser) {
    main.innerHTML = `
      <div class="chat-page fade-in">
        <p class="error-message">You must be logged in to use chat.</p>
      </div>
    `;
    return;
  }

  // Render chat thread UI into main content
  main.innerHTML = `
    <div class="chat-page fade-in">
      <div class="chat-page-header">
        <div>
          <h2>
            Chat with <span id="chatPartnerName">${username}</span>
          </h2>
          <span id="chatPartnerStatus" class="online-status">Offline</span>
        </div>
      </div>

      <div id="chatThreadContainer" class="chat-thread-container">
        <div id="messagesContainer" class="messages-container">
          <div id="messagesWrapper" class="messages-wrapper"></div>
        </div>

        <form id="messageForm" class="message-form">
          <input
            id="messageInput"
            type="text"
            placeholder="Type a message..."
            autocomplete="off"
          />
          <button type="submit" class="btn-accent">Send</button>
        </form>
      </div>
    </div>
  `;

  const messagesContainer = document.getElementById("messagesContainer");
  const messagesWrapper = document.getElementById("messagesWrapper");
  const messageForm = document.getElementById("messageForm");
  const messageInput = document.getElementById("messageInput");
  const statusEl = document.getElementById("chatPartnerStatus");
  const partnerNameEl = document.getElementById("chatPartnerName");

  partnerNameEl.textContent = username;

  // set initial status based on what sidebar knows
  statusEl.textContent = isOnline ? "Online" : "Offline";
  statusEl.className = isOnline ? "online-status online" : "online-status";

  // ====== Local state for pagination ======
  let offset = 0;
  let isLoading = false;
  let hasMore = true;

  // ====== Helpers ======

  function formatDateTime(date) {
    const today = new Date();
    const isToday = date.toDateString() === today.toDateString();
    if (isToday) {
      return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }

  function createMessageElement(msg, currentUserId) {
    const isSent = msg.sender_id === currentUserId;
    const el = document.createElement("div");
    el.className = `message ${isSent ? "sent" : "received"}`;
    el.dataset.messageId = msg.message_id;

    const ts = new Date(msg.created_at);
    el.innerHTML = `
      <div class="message-bubble">${escapeHtml(msg.content)}</div>
      <div class="message-info">
        ${msg.sender_name} • ${formatDateTime(ts)}
      </div>
    `;
    return el;
  }

  function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
  }

  function appendMessage(msg, doScroll = true) {
    const el = createMessageElement(msg, currentUser.id);
    messagesWrapper.appendChild(el);
    if (doScroll) scrollToBottom();
  }

  // ====== Load messages ======

  async function loadMessages(initial = true) {
    if (isLoading || !hasMore) return;
    isLoading = true;

    try {
      const resp = await fetch(
        `${API_BASE}/messages/conversation?user_id=${encodeURIComponent(
          userId
        )}&limit=10&offset=${offset}`,
        { credentials: "include" }
      );

      if (!resp.ok) {
        console.error("Failed to load messages:", resp.status);
        return;
      }

      const data = await resp.json();
      hasMore = data.has_more;
      const messages = (data.messages || []).reverse(); // oldest first

      if (initial) {
        messagesWrapper.innerHTML = "";
        messages.forEach((m) => appendMessage(m, false));
        scrollToBottom();
      } else {
        // prepend older messages while preserving scroll position
        const oldHeight = messagesContainer.scrollHeight;
        messages.forEach((m) => {
          const el = createMessageElement(m, currentUser.id);
          messagesWrapper.insertBefore(el, messagesWrapper.firstChild);
        });
        const newHeight = messagesContainer.scrollHeight;
        messagesContainer.scrollTop = newHeight - oldHeight;
      }

      offset += messages.length;
    } catch (err) {
      console.error("Error loading messages:", err);
    } finally {
      isLoading = false;
    }
  }

  // Infinite scroll up to load older messages
  messagesContainer.addEventListener("scroll", () => {
    if (messagesContainer.scrollTop === 0 && hasMore && !isLoading) {
      loadMessages(false);
    }
  });

  // ====== Send message (with CSRF) ======

  messageForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    const content = messageInput.value.trim();
    if (!content) return;

    const token = getCSRF();
    if (!token) {
      console.warn("[chat] No CSRF token available when sending message");
    }

    try {
      const resp = await fetch(`${API_BASE}/messages/send`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token && { "X-CSRF-Token": token }), // 👈 CSRF header
        },
        credentials: "include",
        body: JSON.stringify({
          receiver_id: userId,
          content,
        }),
      });

      if (!resp.ok) {
        const errData = await resp.json().catch(() => ({}));
        alert(errData.error || "Failed to send message");
        return;
      }

      const data = await resp.json();
      const msg = {
        ...data.message,
        sender_name: currentUser.username,
      };
      appendMessage(msg, true);
      messageInput.value = "";
    } catch (err) {
      console.error("Send message error:", err);
      alert("Failed to send message");
    }
  });

  // ====== Online/offline status hook (used by sidebar via window) ======

  // liveChatSidebar.js can call: window.updateChatPartnerStatus(userId, isOnline)
  window.updateChatPartnerStatus = (partnerId, isOnline) => {
    // Only update if this chat is for that user
    if (String(partnerId) !== String(userId)) return;
    statusEl.textContent = isOnline ? "Online" : "Offline";
    statusEl.className = isOnline ? "online-status online" : "online-status";
  };

  // Receive WebSocket messages from sidebar
  window.receiveChatMessage = (msg) => {
    // Only messages from the active user
    if (String(msg.sender_id) !== String(userId)) return;
    appendMessage(msg, true);
  };

  // ====== Initial load ======
  await loadMessages(true);
}
