document.addEventListener('DOMContentLoaded', () => {

// API Configuration
const API_BASE = 'http://localhost:8080';
const WS_URL = 'ws://localhost:8080/ws';

// State Management
let currentUser = null;
let csrfToken = null;
let ws = null;
let selectedUserId = null;
let conversations = new Map();
let onlineUsers = new Set();
let messageOffset = 0;
let hasMoreMessages = false;
let isLoadingMessages = false;
let isAuthMode = 'login'; // 'login' or 'register'

// DOM Elements
const authContainer = document.getElementById('authContainer');
const chatContainer = document.getElementById('chatContainer');
const authForm = document.getElementById('authForm');
const authMessage = document.getElementById('authMessage');
const conversationsList = document.getElementById('conversationsList');
const messagesContainer = document.getElementById('messagesContainer');
const messagesWrapper = document.getElementById('messagesWrapper');
const messageForm = document.getElementById('messageForm');
const messageInput = document.getElementById('messageInput');
const noConversation = document.getElementById('noConversation');
const chatContent = document.getElementById('chatContent');
const chatHeader = document.getElementById('chatHeader');
const logoutBtn = document.getElementById('logoutBtn');
const newChatBtn = document.getElementById('newChatBtn');
const newChatModal = document.getElementById('newChatModal');
const closeModal = document.getElementById('closeModal');
const usersList = document.getElementById('usersList');

// Safe modal toggle (only if exists)
function openNewChatModal() {
  if (newChatModal) newChatModal.classList.add('active');
}
function closeNewChatModal() {
  if (newChatModal) newChatModal.classList.remove('active');
}

// In initializeChat(), attach listeners only if elements exist and not already attached
if (newChatBtn) {
  newChatBtn.addEventListener('click', openNewChatModal);
}
if (closeModal) {
  closeModal.addEventListener('click', closeNewChatModal);
}
if (newChatModal) {
  newChatModal.addEventListener('click', (e) => {
    if (e.target === newChatModal) closeNewChatModal();
  });
}

// Throttle function for scroll events
function throttle(func, delay) {
    let lastCall = 0;
    return function (...args) {
        const now = Date.now();
        if (now - lastCall >= delay) {
            lastCall = now;
            func(...args);
        }
    };
}

// Debounce function for typing indicators
function debounce(func, delay) {
    let timeout;
    return function (...args) {
        clearTimeout(timeout);
        timeout = setTimeout(() => func(...args), delay);
    };
}

// ============ Authentication ============

// Toggle between login and register
document.getElementById('toggleLink').addEventListener('click', () => {
    isAuthMode = isAuthMode === 'login' ? 'register' : 'login';
    const registerFields = document.getElementById('registerFields');
    const authTitle = document.getElementById('authTitle');
    const authSubmit = document.getElementById('authSubmit');
    const toggleText = document.getElementById('toggleText');
    const toggleLink = document.getElementById('toggleLink');
    const loginField = document.getElementById('loginField');

    if (isAuthMode === 'register') {
        registerFields.style.display = 'block';
        authTitle.textContent = 'Register';
        authSubmit.textContent = 'Register';
        toggleText.textContent = 'Already have an account?';
        toggleLink.textContent = 'Login';
        loginField.placeholder = 'Email';
        loginField.type = 'email';
    } else {
        registerFields.style.display = 'none';
        authTitle.textContent = 'Login';
        authSubmit.textContent = 'Login';
        toggleText.textContent = "Don't have an account?";
        toggleLink.textContent = 'Register';
        loginField.placeholder = 'Username or Email';
        loginField.type = 'text';
    }
    authMessage.style.display = 'none';
});

// Handle auth form submission
authForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    if (isAuthMode === 'login') {
        await handleLogin();
    } else {
        await handleRegister();
    }
});

async function handleLogin() {
    const login = document.getElementById('loginField').value.trim();
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_BASE}/forum/api/session/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ login, password })
        });

        const data = await response.json();

        if (response.ok) {
            currentUser = data.user;
            csrfToken = data.csrf_token;
            showMessage('Login successful!', 'success');
            setTimeout(() => initializeChat(), 500);
        } else {
            showMessage(data.error || 'Login failed', 'error');
        }
    } catch (error) {
        showMessage('Error connecting to server', 'error');
        console.error('Login error:', error);
    }
}

async function handleRegister() {
    const username = document.getElementById('username').value.trim();
    const email = document.getElementById('loginField').value.trim();
    const password = document.getElementById('password').value;
    const firstName = document.getElementById('firstName').value.trim();
    const lastName = document.getElementById('lastName').value.trim();
    const age = parseInt(document.getElementById('age').value);
    const gender = document.getElementById('gender').value;

    try {
        const response = await fetch(`${API_BASE}/forum/api/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({
                username,
                email,
                password,
                first_name: firstName,
                last_name: lastName,
                age,
                gender
            })
        });

        const data = await response.json();

        if (response.ok) {
            currentUser = data.user;
            csrfToken = data.csrf_token;
            showMessage('Registration successful!', 'success');
            setTimeout(() => initializeChat(), 500);
        } else {
            showMessage(data.error || 'Registration failed', 'error');
        }
    } catch (error) {
        showMessage('Error connecting to server', 'error');
        console.error('Register error:', error);
    }
}

function showMessage(message, type) {
    authMessage.textContent = message;
    authMessage.className = type === 'success' ? 'success-message' : 'error-message';
    authMessage.style.display = 'block';
}

// ============ Chat Initialization ============

async function initializeChat() {
    authContainer.style.display = 'none';
    chatContainer.style.display = 'grid';
    
    document.getElementById('currentUserInfo').textContent = `Logged in as ${currentUser.username}`;
    
    // Load conversations
    await loadConversations();
    
    // Load all users for new chat
    await loadAllUsers();
    
    // Connect WebSocket
    connectWebSocket();
    
    // // Setup modal handlers
    // newChatBtn.addEventListener('click', () => {
    //     newChatModal.classList.add('active');
    // });
    
    closeModal.addEventListener('click', () => {
        newChatModal.classList.remove('active');
    });
    
    // Close modal when clicking outside
    newChatModal.addEventListener('click', (e) => {
        if (e.target === newChatModal) {
            newChatModal.classList.remove('active');
        }
    });
}

// ============ WebSocket ============

function connectWebSocket() {
    ws = new WebSocket(WS_URL);

    ws.onopen = () => {
        console.log('WebSocket connected');
    };

    ws.onmessage = (event) => {
        const message = JSON.parse(event.data);
        handleWebSocketMessage(message);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
        console.log('WebSocket closed, reconnecting...');
        setTimeout(connectWebSocket, 3000);
    };
}

function handleWebSocketMessage(message) {
    switch (message.type) {
        case 'chat':
            handleIncomingMessage(message.data);
            break;
        case 'online_status':
            handleOnlineStatus(message.data);
            break;
        case 'online_users_list':
            handleOnlineUsersList(message.data);
            break;
        case 'message_delete':
            handleMessageDelete(message.data);
            break;
        case 'typing':
            handleTypingIndicator(message.data);
            break;
    }
}

function handleIncomingMessage(data) {
    // Update conversation in sidebar
    updateConversationLastMessage(data.sender_id, data.content, new Date(data.created_at));
    
    // If this conversation is currently selected, append message
    if (selectedUserId === data.sender_id) {
        appendMessage(data, false);
        scrollToBottom();
    }
    
    // Play notification sound or show notification
    if (selectedUserId !== data.sender_id) {
        // Increment unread count
        const convElement = document.querySelector(`[data-user-id="${data.sender_id}"]`);
        if (convElement) {
            let badge = convElement.querySelector('.unread-badge');
            if (!badge) {
                badge = document.createElement('div');
                badge.className = 'unread-badge';
                badge.textContent = '1';
                convElement.appendChild(badge);
            } else {
                badge.textContent = parseInt(badge.textContent) + 1;
            }
        }
    }
}

function handleOnlineStatus(data) {
    if (data.is_online) {
        onlineUsers.add(data.user_id);
    } else {
        onlineUsers.delete(data.user_id);
    }
    
    // Update UI
    updateUserOnlineStatus(data.user_id, data.is_online);
    
    // Refresh the users list in the modal if it's open
    if (newChatModal.classList.contains('active')) {
        loadAllUsers();
    }
}

function handleOnlineUsersList(users) {
    users.forEach(user => {
        onlineUsers.add(user.user_id);
        updateUserOnlineStatus(user.user_id, true);
    });
    
    // Refresh the users list in the modal if it's open
    if (newChatModal.classList.contains('active')) {
        loadAllUsers();
    }
}

function handleMessageDelete(data) {
    const messageElement = document.querySelector(`[data-message-id="${data.message_id}"]`);
    if (messageElement) {
        messageElement.remove();
    }
}

function handleTypingIndicator(data) {
    // Implement typing indicator if needed
    console.log(`${data.username} is typing...`);
}

// ============ Conversations ============

async function loadConversations() {
    try {
        const response = await fetch(`${API_BASE}/forum/api/messages/conversations`, {
            credentials: 'include'
        });

        if (response.ok) {
            const data = await response.json();
            renderConversations(data.conversations || []);
        }
    } catch (error) {
        console.error('Failed to load conversations:', error);
    }
}

async function loadAllUsers() {
    try {
        // ✅ FIX: Call the correct endpoint for starting new chats
        const response = await fetch(`${API_BASE}/forum/api/messages/users-for-chat?all=true`, {
            credentials: 'include'
        });

        if (response.ok) {
            const data = await response.json();
            renderUsersList(data.users || []);
        }
    } catch (error) {
        console.error('Failed to load users:', error);
    }
}

function renderUsersList(users) {
    usersList.innerHTML = '';
    
    if (users.length === 0) {
        usersList.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">No users available</div>';
        return;
    }

    // Separate online and offline users, then sort alphabetically
    const onlineUsersList = users.filter(u => onlineUsers.has(u.id)).sort((a, b) => a.username.localeCompare(b.username));
    const offlineUsersList = users.filter(u => !onlineUsers.has(u.id)).sort((a, b) => a.username.localeCompare(b.username));
    
    // Render online users first
    onlineUsersList.forEach(user => {
        const item = createUserListItem(user, true);
        usersList.appendChild(item);
    });
    
    // Then offline users
    offlineUsersList.forEach(user => {
        const item = createUserListItem(user, false);
        usersList.appendChild(item);
    });
}

function createUserListItem(user, isOnline) {
    const item = document.createElement('div');
    item.className = 'user-item';
    
    item.innerHTML = `
        <div class="online-indicator ${isOnline ? 'online' : ''}"></div>
        <div class="user-details">
            <div class="user-name">${user.username}</div>
            <div class="user-email">${user.email}</div>
        </div>
    `;
    
    item.addEventListener('click', () => {
        newChatModal.classList.remove('active');
        startConversationWithUser(user.id, user.username);
    });
    
    return item;
}

function startConversationWithUser(userId, username) {
    // Check if conversation already exists
    const existingConv = document.querySelector(`[data-user-id="${userId}"]`);
    if (existingConv) {
        existingConv.click();
        return;
    }
    
    // Create a new conversation item
    const conv = {
        user_id: userId,
        username: username,
        last_message: 'Start chatting...',
        last_message_time: new Date().toISOString(),
        unread_count: 0,
        is_online: onlineUsers.has(userId)
    };
    
    const item = createConversationElement(conv);
    conversationsList.insertBefore(item, conversationsList.firstChild);
    
    // Select the conversation
    selectConversation(userId, username);
}

function renderConversations(convs) {
    conversationsList.innerHTML = '';
    
    if (!convs || convs.length === 0) {
        conversationsList.innerHTML = '<div style="padding: 20px; text-align: center; color: #999;">No conversations yet<br><small>Click "New Chat" to start</small></div>';
        return;
    }

    // Sort by last message time (most recent first)
    convs.sort((a, b) => new Date(b.last_message_time) - new Date(a.last_message_time));

    convs.forEach(conv => {
        const item = createConversationElement(conv);
        conversationsList.appendChild(item);
    });
}

function createConversationElement(conv) {
    const item = document.createElement('div');
    item.className = 'conversation-item';
    item.dataset.userId = conv.user_id;
    
    const isOnline = conv.is_online || onlineUsers.has(conv.user_id);
    
    item.innerHTML = `
        <div class="user-name">
            <div class="online-indicator ${isOnline ? 'online' : ''}"></div>
            ${conv.username}
        </div>
        <div class="last-message">${conv.last_message || 'No messages yet'}</div>
        ${conv.unread_count > 0 ? `<div class="unread-badge">${conv.unread_count}</div>` : ''}
        <div class="timestamp">${formatTime(new Date(conv.last_message_time))}</div>
    `;
    
    item.addEventListener('click', () => selectConversation(conv.user_id, conv.username));
    
    return item;
}

function updateConversationLastMessage(userId, message, timestamp) {
    const item = document.querySelector(`[data-user-id="${userId}"]`);
    if (item) {
        item.querySelector('.last-message').textContent = message;
        item.querySelector('.timestamp').textContent = formatTime(timestamp);
        
        // Move to top
        conversationsList.insertBefore(item, conversationsList.firstChild);
    } else {
        // Reload conversations to get the new one
        loadConversations();
    }
}

function updateUserOnlineStatus(userId, isOnline) {
    const item = document.querySelector(`[data-user-id="${userId}"]`);
    if (item) {
        const indicator = item.querySelector('.online-indicator');
        if (isOnline) {
            indicator.classList.add('online');
        } else {
            indicator.classList.remove('online');
        }
    }
    
    // Update chat header if this is the selected user
    if (selectedUserId === userId) {
        const statusElement = document.getElementById('selectedUserStatus');
        statusElement.textContent = isOnline ? 'Online' : 'Offline';
        statusElement.className = isOnline ? 'online-status online' : 'online-status';
    }
}

// ============ Messages ============

async function selectConversation(userId, username) {
    selectedUserId = userId;
    messageOffset = 0;
    hasMoreMessages = false;
    
    // Update UI
    document.querySelectorAll('.conversation-item').forEach(item => {
        item.classList.remove('active');
    });
    document.querySelector(`[data-user-id="${userId}"]`).classList.add('active');
    
    // Clear unread badge
    const badge = document.querySelector(`[data-user-id="${userId}"] .unread-badge`);
    if (badge) {
        badge.remove();
    }
    
    // Show chat area
    noConversation.style.display = 'none';
    chatContent.style.display = 'flex';
    chatHeader.style.display = 'flex';
    
    // Update header
    document.getElementById('selectedUserName').textContent = username;
    const isOnline = onlineUsers.has(userId);
    const statusElement = document.getElementById('selectedUserStatus');
    statusElement.textContent = isOnline ? 'Online' : 'Offline';
    statusElement.className = isOnline ? 'online-status online' : 'online-status';
    
    // Load messages
    await loadMessages(userId, true);
    
    // Setup scroll listener for pagination
    messagesContainer.removeEventListener('scroll', handleScroll);
    messagesContainer.addEventListener('scroll', throttle(handleScroll, 200));
}

async function loadMessages(userId, initialLoad = false) {
    if (isLoadingMessages) return;
    isLoadingMessages = true;
    
    try {
        const response = await fetch(
            `${API_BASE}/forum/api/messages/conversation?user_id=${userId}&limit=10&offset=${messageOffset}`,
            { credentials: 'include' }
        );

        if (response.ok) {
            const data = await response.json();
            hasMoreMessages = data.has_more;
            
            if (initialLoad) {
                messagesWrapper.innerHTML = '';
            }
            
            // Messages come in reverse order (newest first), so we reverse to display oldest first
            const messages = data.messages.reverse();
            
            if (!initialLoad) {
                // When loading more, prepend to top
                messages.forEach(msg => prependMessage(msg));
            } else {
                // Initial load, append normally
                messages.forEach(msg => appendMessage(msg));
                scrollToBottom();
            }
            
            messageOffset += messages.length;
        }
    } catch (error) {
        console.error('Failed to load messages:', error);
    } finally {
        isLoadingMessages = false;
    }
}

function appendMessage(data, scrollDown = true) {
    const isSent = data.sender_id === currentUser.id;
    const message = createMessageElement(data, isSent);
    messagesWrapper.appendChild(message);
    if (scrollDown) {
        scrollToBottom();
    }
}

function prependMessage(data) {
    const isSent = data.sender_id === currentUser.id;
    const message = createMessageElement(data, isSent);
    messagesWrapper.insertBefore(message, messagesWrapper.firstChild);
}

function createMessageElement(data, isSent) {
    const message = document.createElement('div');
    message.className = `message ${isSent ? 'sent' : 'received'}`;
    message.dataset.messageId = data.message_id;
    
    const timestamp = new Date(data.created_at);
    const timeStr = formatDateTime(timestamp);
    
    message.innerHTML = `
        <div class="message-bubble">${escapeHtml(data.content)}</div>
        <div class="message-info">${data.sender_name} • ${timeStr}</div>
    `;
    
    return message;
}

// Handle scroll for loading more messages
function handleScroll() {
    // Check if scrolled to top
    if (messagesContainer.scrollTop === 0 && hasMoreMessages && !isLoadingMessages) {
        const scrollHeight = messagesContainer.scrollHeight;
        loadMessages(selectedUserId, false).then(() => {
            // Maintain scroll position
            const newScrollHeight = messagesContainer.scrollHeight;
            messagesContainer.scrollTop = newScrollHeight - scrollHeight;
        });
    }
}

function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// ============ Send Message ============

messageForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const content = messageInput.value.trim();
    if (!content || !selectedUserId) return;
    
    try {
        const response = await fetch(`${API_BASE}/forum/api/messages/send`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken
            },
            credentials: 'include',
            body: JSON.stringify({
                receiver_id: selectedUserId,
                content: content
            })
        });

        if (response.ok) {
            const data = await response.json();
            // Message will be shown when WebSocket broadcasts it back
            // Or append immediately
            const msgData = {
                ...data.message,
                sender_name: currentUser.username
            };
            appendMessage(msgData);
            updateConversationLastMessage(selectedUserId, content, new Date());
            
            messageInput.value = '';
        } else {
            const error = await response.json();
            alert(error.error || 'Failed to send message');
        }
    } catch (error) {
        console.error('Failed to send message:', error);
        alert('Failed to send message');
    }
});

// ============ Logout ============

logoutBtn.addEventListener('click', async () => {
    try {
        await fetch(`${API_BASE}/forum/api/session/logout`, {
            method: 'POST',
            credentials: 'include'
        });
        
        if (ws) {
            ws.close();
        }
        
        // Reload page
        window.location.reload();
    } catch (error) {
        console.error('Logout error:', error);
    }
});

// ============ Utility Functions ============

function formatTime(date) {
    const now = new Date();
    const diff = now - date;
    
    if (diff < 60000) { // Less than 1 minute
        return 'Just now';
    } else if (diff < 3600000) { // Less than 1 hour
        const mins = Math.floor(diff / 60000);
        return `${mins}m ago`;
    } else if (diff < 86400000) { // Less than 1 day
        const hours = Math.floor(diff / 3600000);
        return `${hours}h ago`;
    } else {
        return date.toLocaleDateString();
    }
}

function formatDateTime(date) {
    const today = new Date();
    const isToday = date.toDateString() === today.toDateString();
    
    if (isToday) {
        return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
    } else {
        return date.toLocaleString('en-US', { 
            month: 'short', 
            day: 'numeric', 
            hour: '2-digit', 
            minute: '2-digit' 
        });
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ============ Initialize ============

// Check if already logged in
async function checkSession() {
    try {
        const response = await fetch(`${API_BASE}/forum/api/session/verify`, {
            credentials: 'include'
        });

        if (response.ok) {
            const data = await response.json();
            currentUser = data.user;
            csrfToken = data.csrf_token;
            initializeChat();
        }
    } catch (error) {
        console.log('No active session');
    }
}

// Check session on load
checkSession();
});