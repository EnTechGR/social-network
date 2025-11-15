import { navigateTo } from "../router.js";

export function renderWelcome(app) {
  app.innerHTML = `
    <section class="welcome-page fade-in">
      <h1>Welcome to BookTalk</h1>
      <p>Your place to discuss stories, share reviews, and explore new worlds — one page at a time.</p>

      <div class="welcome-actions">
        <button id="loginBtn">Login</button>
        <button id="registerBtn">Register</button>
      </div>
    </section>
  `;

  // Navigation buttons
  document.getElementById("loginBtn").addEventListener("click", () => {
    navigateTo("/login");
  });

  document.getElementById("registerBtn").addEventListener("click", () => {
    navigateTo("/register");
  });
}
