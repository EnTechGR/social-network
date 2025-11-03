import { verifySession } from "./session.js";
import { router } from "./router.js";

async function init() {
  const app = document.getElementById("app");

  // Show a minimal loader during startup
  app.innerHTML = `<div class="loader">Loading...</div>`;

  try {
    // Verify session (user or guest)
    await verifySession();
  } catch (err) {
    console.warn("Session verification failed:", err);
  }

  // Render initial route
  router();
}

// Handle browser back/forward buttons
window.addEventListener("popstate", () => {
  router();
});

// Initialize the SPA
init();
