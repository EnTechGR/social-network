const bell = document.getElementById('notification-bell');
const modal = document.getElementById('notifications-modal');
const closeBtn = modal ? modal.querySelector('.close-btn') : null;
const listContainer = document.getElementById('notification-list');
const template = document.getElementById('notif-item-template');

// CSRF token stored after verifying the session
let csrfTokenFromResponse = null;
const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';

// Load CSRF token from the current session
async function loadCSRFTokenFromSession() {
  try {
    const resp = await fetch(sessionVerifyURL, { credentials: 'include' });
    if (!resp.ok) throw new Error('Session not valid');
    const data = await resp.json();
    return data.csrf_token || data.CSRFToken;
  } catch (err) {
    console.warn('Failed to load CSRF token from session:', err);
    return null;
  }
}

// Fetch CSRF token when the page loads
window.addEventListener('DOMContentLoaded', async () => {
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
});

async function loadNotifications() {
  try {
    const resp = await fetch('http://localhost:8080/forum/api/notifications', { credentials: 'include' });
    if (!resp.ok) throw new Error('Failed to load notifications');
    const data = await resp.json();
    const items = Array.isArray(data.notifications) ? data.notifications : [];
    renderNotifications(items);
  } catch (err) {
    console.error('Error loading notifications:', err);
    renderNotifications([]);
  }
}

function renderNotifications(items) {
  if (!listContainer) return;
  listContainer.innerHTML = '';
  if (!items.length) {
    listContainer.textContent = 'No notifications';
    return;
  }
  const frag = document.createDocumentFragment();
  items.forEach((n) => {
    const node = template.content.cloneNode(true);
    node.querySelector('.notification-message').textContent = formatMessage(n);
    if (n.created_at) {
      node.querySelector('.notification-time').textContent = new Date(n.created_at).toLocaleString();
    }
    const delBtn = node.querySelector('.delete-notification-btn');
    if (delBtn) {
      delBtn.addEventListener('click', async () => {
        try {
      if (!csrfTokenFromResponse) {
        csrfTokenFromResponse = await loadCSRFTokenFromSession();
      }
      await fetch(`http://localhost:8080/forum/api/notifications/delete/${n.id}`, {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          'X-CSRF-Token': csrfTokenFromResponse,
        },
      });
      delBtn.parentElement.remove();
      loadNotifications();
        } catch (err) {
          console.error('Failed to delete notification', err);
        }
      });
    }
    frag.appendChild(node);
  });
  listContainer.appendChild(frag);
}

function formatMessage(n) {
  const actor = n.username || 'Someone';
  const onComment = n.comment_id !== undefined && n.comment_id !== null;
  switch (n.type) {
    case 'like':
      return `${actor} liked your ${onComment ? 'comment' : 'post'}`;
    case 'dislike':
      return `${actor} disliked your ${onComment ? 'comment' : 'post'}`;
    case 'comment':
      return `${actor} commented on your post`;
    case 'edit_comment':
      return `${actor} edited a comment on your post`;
    case 'delete_comment':
      return `${actor} deleted a comment on your post`;
    default:
      return `${actor} interacted with your post`;
  }
}

if (bell && modal && closeBtn) {
  bell.addEventListener('click', async (e) => {
    e.preventDefault();
    modal.classList.remove('hidden');
    await loadNotifications();
  });

  closeBtn.addEventListener('click', () => {
    modal.classList.add('hidden');
  });

  window.addEventListener('click', (e) => {
    if (e.target === modal) {
      modal.classList.add('hidden');
    }
  });
}
