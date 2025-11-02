import { verifySession } from "./session.js";
import { router } from "./router.js";

async function init() {
  await verifySession();
  router();
}

// Handle back/forward navigation
window.addEventListener("popstate", () => {
  router();
});

init();
