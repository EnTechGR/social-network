// /static/views/notifications.js
import { ensureSessionChecked, getUser, getCSRF } from "../session.js";
import { navigateTo } from "../router.js";

const API_BASE = "http://localhost:8080/forum/api";

export async function renderNotifications(main) {
  // Make sure we know if user is logged in and have CSRF
  await ensureSessionChecked();
  const user = getUser();

  if (!user) {
    return navigateTo("/login");
  }

  // Initial layout
  main.innerHTML = `
    <section class="notifications-page fade-in">
      <header class="notifications-header">
        <h2>Notifications</h2>
        <p id="notificationsCount" class="notifications-count"></p>
      </header>

      <div id="notificationsContainer" class="notifications-container">
        <div class="loader">Loading your notifications...</div>
      </div>
    </section>
  `;

  const container = document.getElementById("notificationsContainer");
  const countEl = document.getElementById("notificationsCount");

  async function loadNotifications() {
    try {
      const resp = await fetch(`${API_BASE}/notifications`, {
        credentials: "include",
      });

      if (!resp.ok) throw new Error("Failed to load notifications");

      const data = await resp.json();
      const items = Array.isArray(data.notifications) ? data.notifications : [];

      if (typeof data.count === "number") {
        countEl.textContent =
          data.count === 0
            ? "You have no notifications"
            : `You have ${data.count} notification${
                data.count === 1 ? "" : "s"
              }`;
      } else {
        countEl.textContent = "";
      }

      renderNotifications(items);
    } catch (err) {
      console.error("Error loading notifications:", err);
      countEl.textContent = "";
      renderNotifications([]);
    }
  }

  function renderNotifications(items) {
    if (!container) return;

    if (!items.length) {
      container.innerHTML = `
        <p class="empty-state">
          No notifications yet. Interact with posts to see updates here.
        </p>
      `;
      return;
    }

    container.innerHTML = "";

    const fragment = document.createDocumentFragment();

    items.forEach((n) => {
      const item = document.createElement("article");
      item.className = "notification-item";

      const messageEl = document.createElement("div");
      messageEl.className = "notification-message";
      messageEl.textContent = formatMessage(n);

      const metaEl = document.createElement("div");
      metaEl.className = "notification-meta";

      if (n.created_at) {
        const timeEl = document.createElement("span");
        timeEl.className = "notification-time";
        timeEl.textContent = new Date(n.created_at).toLocaleString();
        metaEl.appendChild(timeEl);
      }

      const actionsEl = document.createElement("div");
      actionsEl.className = "notification-actions";

      // Optional "View" / link button if backend sends a link
      if (n.link) {
        const viewBtn = document.createElement("button");
        viewBtn.type = "button";
        viewBtn.className = "btn-link";
        viewBtn.textContent = "View";

        viewBtn.addEventListener("click", (e) => {
          e.preventDefault();
          // If it's an internal app route, SPA-nav; otherwise, full redirect
          if (n.link.startsWith("/")) {
            navigateTo(n.link);
          } else {
            window.location.href = n.link;
          }
        });

        actionsEl.appendChild(viewBtn);
      }

      const deleteBtn = document.createElement("button");
      deleteBtn.type = "button";
      deleteBtn.className = "btn-link danger delete-notification-btn";
      deleteBtn.textContent = "Dismiss";

      deleteBtn.addEventListener("click", async (e) => {
        e.preventDefault();
        await deleteNotification(n.id);
      });

      actionsEl.appendChild(deleteBtn);

      item.appendChild(messageEl);
      item.appendChild(metaEl);
      item.appendChild(actionsEl);

      fragment.appendChild(item);
    });

    container.appendChild(fragment);
  }

  async function deleteNotification(id) {
    try {
      const token = getCSRF();
      if (!token) {
        console.warn("[notifications] No CSRF token found");
      }

      const resp = await fetch(`${API_BASE}/notifications/delete/${id}`, {
        method: "DELETE",
        credentials: "include",
        headers: {
          ...(token && { "X-CSRF-Token": token }),
        },
      });

      if (!resp.ok) {
        const errData = await resp.json().catch(() => ({}));
        console.error("Failed to delete notification", errData);
        return;
      }

      // Reload after delete
      await loadNotifications();
    } catch (err) {
      console.error("Failed to delete notification", err);
    }
  }

  function formatMessage(n) {
    const actor = n.username || "Someone";
    const onComment = n.comment_id !== undefined && n.comment_id !== null;
    const onPost = n.post_id !== undefined && n.post_id !== null;

    switch (n.type) {
      case "like":
        if (onComment) return `${actor} liked your comment`;
        if (onPost) return `${actor} liked your post`;
        return `${actor} liked your content`;
      case "dislike":
        if (onComment) return `${actor} disliked your comment`;
        if (onPost) return `${actor} disliked your post`;
        return `${actor} disliked your content`;
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

  // 🔄 initial load
  loadNotifications();
}
