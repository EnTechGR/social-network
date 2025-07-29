const container = document.getElementById('notificationsContainer');
const countEl = document.getElementById('notificationsCount');
const template = document.getElementById('notif-item-template');

// Store CSRF token from the session
let csrfTokenFromResponse = null;
const sessionVerifyURL = 'http://localhost:8080/forum/api/session/verify';

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

window.addEventListener('DOMContentLoaded', async () => {
  csrfTokenFromResponse = await loadCSRFTokenFromSession();
  loadNotifications();
});

async function loadNotifications() {
  try {
    // Correct API path per backend routes
    const resp = await fetch('http://localhost:8080/forum/api/notifications', { credentials: 'include' });
    if (!resp.ok) throw new Error('Failed to load notifications');
    const data = await resp.json();
    if (data && typeof data.count === 'number') {
      countEl.textContent = `You have ${data.count} notifications`;
    }
    const list = Array.isArray(data.notifications) ? data.notifications : [];
    renderNotifications(list);
  } catch (err) {
    console.error('Error loading notifications:', err);
    renderNotifications([]);
  }
}

function renderNotifications(items) {
  if (!container) {
    console.error('Notifications container element not found');
    return;
  }
  if (!template) {
    console.error('Notification template element not found');
    container.textContent = 'Error: Notification template missing.';
    return;
  }
  container.innerHTML = '';
  if (!items.length) {
    container.textContent = 'No notifications yet.';
    return;
  }
  const fragment = document.createDocumentFragment();
  items.forEach(n => {
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
    if (n.link) {
      const wrapper = document.createElement('a');
      wrapper.href = n.link;
      wrapper.className = 'notification-link';
      wrapper.appendChild(node);
      fragment.appendChild(wrapper);
    } else {
      fragment.appendChild(node);
    }
  });
  container.appendChild(fragment);
}

function formatMessage(n) {
  const actor = n.username || 'Someone';
  const onComment = n.comment_id !== undefined && n.comment_id !== null;
  const onPost = n.post_id !== undefined && n.post_id !== null;
  switch (n.type) {
    case 'like':
      if (onComment) return `${actor} liked your comment`;
      if (onPost) return `${actor} liked your post`;
      return `${actor} liked your content`;
    case 'dislike':
      if (onComment) return `${actor} disliked your comment`;
      if (onPost) return `${actor} disliked your post`;
      return `${actor} disliked your content`;
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
