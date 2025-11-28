// /static/renderChat.js - FIXED VERSION

import { getCSRF, ensureSessionChecked, getUser } from "../session.js";

const API_BASE = "http://localhost:8080/forum/api";
const IMAGE_SERVE_BASE = `${API_BASE}/chat/images/serve`;

export async function renderChat(main, { userId, username, isOnline = false }) {
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
          <div id="imagePreviewContainer" class="image-preview-container">
            </div>
          <div class="input-row">
            <label for="imageUploadInput" class="image-upload-label" title="Send an image">
              <span class="icon-image">🖼️</span> 
            </label>
            <input
              id="imageUploadInput"
              type="file"
              accept="image/*"
              style="display: none;"
            />
            <input
              id="messageInput"
              type="text"
              placeholder="Type a message or image caption..."
              autocomplete="off"
            />
            <button type="submit" class="btn-accent" id="sendMessageBtn">Send</button>
          </div>
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
  
  const imageUploadInput = document.getElementById("imageUploadInput");
  const sendMessageBtn = document.getElementById("sendMessageBtn");
  const imagePreviewContainer = document.getElementById("imagePreviewContainer"); 

  partnerNameEl.textContent = username;

  statusEl.textContent = isOnline ? "Online" : "Offline";
  statusEl.className = isOnline ? "online-status online" : "online-status";

  // ====== Local state ======
  let selectedFile = null;
  let offset = 0;
  let isLoading = false;
  let hasMore = true;

  // ====== Helpers ======

  function throttle(fn, wait) {
    let lastTime = 0;
    let timeoutId = null;

    return function (...args) {
      const now = Date.now();
      const remaining = wait - (now - lastTime);

      if (remaining <= 0) {
        if (timeoutId) {
          clearTimeout(timeoutId);
          timeoutId = null;
        }
        lastTime = now;
        fn.apply(this, args);
      } else if (!timeoutId) {
        timeoutId = setTimeout(() => {
          lastTime = Date.now();
          timeoutId = null;
          fn.apply(this, args);
        }, remaining);
      }
    };
  }

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

  // ✅ FIXED: Better image handling with debugging
  function createMessageElement(msg, currentUserId) {
    const isSent = msg.sender_id === currentUserId;
    const el = document.createElement("div");
    el.className = `message ${isSent ? "sent" : "received"}`;
    el.dataset.messageId = msg.message_id;

    const ts = new Date(msg.created_at);
    let contentHTML = '';

    // ✅ DEBUG: Log the message to see what we're receiving
    console.log('Creating message element:', {
      message_id: msg.message_id,
      has_image: !!msg.image,
      image_data: msg.image
    });

    // ✅ FIXED: Check for image data with better null checking
    if (msg.image && msg.image.file_path) {
        const imagePath = `${IMAGE_SERVE_BASE}/${msg.image.file_path}`;
        const imageThumbnailPath = `${IMAGE_SERVE_BASE}/${msg.image.file_path}`;

        contentHTML = `
            <div class="message-bubble image-message">
                <a href="${imagePath}" target="_blank">
                    <img 
                        src="${imageThumbnailPath}" 
                        alt="Chat Image" 
                        loading="lazy"
                        class="chat-image-thumbnail"
                        onerror="console.error('Failed to load image: ${imagePath}')"
                    />
                </a>
                ${msg.content ? `<p class="image-caption">${escapeHtml(msg.content)}</p>` : ''}
            </div>
        `;
    } else {
      // Regular text message
      contentHTML = `
        <div class="message-bubble">${escapeHtml(msg.content)}</div>
      `;
    }

    el.innerHTML = `
        ${contentHTML}
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
    // ✅ FIXED: Check for duplicate messages before appending
    const existingMessage = messagesWrapper.querySelector(`[data-message-id="${msg.message_id}"]`);
    if (existingMessage) {
      console.log('Message already exists, skipping:', msg.message_id);
      return;
    }

    const el = createMessageElement(msg, currentUser.id);
    messagesWrapper.appendChild(el);
    if (doScroll) scrollToBottom();
  }
  
  // ====== Image Preview Logic ======

  function clearImagePreview() {
    imagePreviewContainer.innerHTML = '';
    imagePreviewContainer.style.display = 'none';
    imageUploadInput.value = '';
    selectedFile = null;
    messageInput.placeholder = "Type a message or image caption...";
  }

  function showImagePreview(file) {
    const reader = new FileReader();
    reader.onload = (e) => {
      const previewHTML = `
        <div class="preview-card">
          <img src="${e.target.result}" alt="Image Preview" class="preview-image"/>
          <button type="button" class="delete-preview-btn">❌</button>
        </div>
      `;
      imagePreviewContainer.innerHTML = previewHTML;
      imagePreviewContainer.style.display = 'block';
      messageInput.placeholder = "Add a caption (optional)...";

      imagePreviewContainer.querySelector('.delete-preview-btn').addEventListener('click', clearImagePreview);
    };
    reader.readAsDataURL(file);
  }

  imageUploadInput.addEventListener('change', (e) => {
    const file = e.target.files[0];
    if (!file) {
      clearImagePreview();
      return;
    }

    const MAX_SIZE = 5 * 1024 * 1024; // 5MB limit
    if (file.size > MAX_SIZE) {
      alert("Image file must be less than 5MB.");
      clearImagePreview();
      return;
    }
    
    selectedFile = file;
    showImagePreview(file);
  });
  
  // ====== Send Message/Image Handler ======

  async function uploadImageAndCaption() {
    sendMessageBtn.disabled = true;
    const file = selectedFile;

    const token = getCSRF();

    const formData = new FormData();
    formData.append('chat_image', file);
    formData.append('receiver_id', userId);

    const caption = messageInput.value.trim();
    if (caption) {
        formData.append('content', caption);
    }

    try {
        const resp = await fetch(`${API_BASE}/chat/images/upload`, {
            method: "POST",
            headers: {
                ...(token && { "X-CSRF-Token": token }),
            },
            credentials: "include",
            body: formData,
        });

        if (!resp.ok) {
            const errData = await resp.json().catch(() => ({}));
            alert(errData.error || "Failed to upload image");
            return;
        }

        const data = await resp.json();
        console.log('Image upload response:', data); // ✅ DEBUG

        const msg = {
            ...data.message,
            sender_name: currentUser.username,
        };

        // ✅ FIXED: Append the message immediately for the SENDER
        appendMessage(msg, true);

        // Clear inputs and preview
        clearImagePreview();
        messageInput.value = '';

    } catch (err) {
        console.error("Upload image error:", err);
        alert("Failed to upload image");
    } finally {
        sendMessageBtn.disabled = false;
    }
  }

  messageForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    
    if (selectedFile) {
      await uploadImageAndCaption();
      return;
    }
    
    const content = messageInput.value.trim();
    if (!content) return;

    const token = getCSRF();
    if (!token) {
      console.warn("[chat] No CSRF token available when sending message");
    }
    
    sendMessageBtn.disabled = true;

    try {
      const resp = await fetch(`${API_BASE}/messages/send`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(token && { "X-CSRF-Token": token }), 
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
    } finally {
        sendMessageBtn.disabled = false;
    }
  });


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
      const messages = (data.messages || []).reverse();

      console.log('Loaded messages:', messages); // ✅ DEBUG

      if (initial) {
        messagesWrapper.innerHTML = "";
        messages.forEach((m) => appendMessage(m, false));
        scrollToBottom();
      } else {
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

  // ====== Infinite scroll ======
  const handleScroll = throttle(() => {
    if (messagesContainer.scrollTop === 0 && hasMore && !isLoading) {
      loadMessages(false);
    }
  }, 250); 

  messagesContainer.addEventListener("scroll", handleScroll);

  // ====== Online/offline status ======

  window.updateChatPartnerStatus = (partnerId, isOnline) => {
    if (String(partnerId) !== String(userId)) return;
    statusEl.textContent = isOnline ? "Online" : "Offline";
    statusEl.className = isOnline ? "online-status online" : "online-status";
  };

  // ✅ FIXED: Better WebSocket message handling with debugging
  window.receiveChatMessage = (rawMsg) => {
    console.log('Received WebSocket message:', rawMsg); // ✅ DEBUG
    
    // Only append if the message is from the user we are currently chatting with
    if (String(rawMsg.sender_id) !== String(userId)) {
      console.log('Message not from current chat partner, ignoring');
      return;
    }

    // ✅ FIXED: Normalize the message object
    const msg = {
        message_id: rawMsg.message_id,
        sender_id: rawMsg.sender_id,
        receiver_id: rawMsg.receiver_id,
        content: rawMsg.content || '',
        created_at: rawMsg.created_at,
        is_read: rawMsg.is_read,
        sender_name: rawMsg.sender_name || username,
        // ✅ CRITICAL: Ensure image is properly passed through
        image: rawMsg.image || null
    };

    console.log('Normalized message for display:', msg); // ✅ DEBUG
    
    appendMessage(msg, true);
  };

  // ====== Cleanup ======
  const cleanup = () => {
    messagesContainer.removeEventListener("scroll", handleScroll);
    if (window.clearActiveChatUser) {
      window.clearActiveChatUser();
    }
  };

  window.cleanupChat = cleanup;

  // ====== Initial load ======
  await loadMessages(true);
}