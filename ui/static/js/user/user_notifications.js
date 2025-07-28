const container = document.getElementById('notificationsContainer');
const countEl = document.getElementById('notificationsCount');
const template = document.getElementById('notification-template');

window.addEventListener('DOMContentLoaded', () => {
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
  switch (n.type) {
    case 'like':
      return `${actor} liked your post`;
    case 'dislike':
      return `${actor} disliked your post`;
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
