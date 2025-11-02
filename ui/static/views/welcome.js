import { navigateTo } from "../router.js";

export function renderWelcome(app) {
  app.innerHTML = `
    <section class="welcome-page">
      <h1>Welcome to the Forum</h1>
      <p>Join discussions, share your thoughts, and connect with others.</p>

      <div class="welcome-actions">
        <button id="loginBtn">Login</button>
        <button id="registerBtn">Register</button>
        <button id="guestBtn">Continue as Guest</button>
      </div>
    </section>
  `;

  document.getElementById("loginBtn").addEventListener("click", () => {
    navigateTo("/login");
  });

  document.getElementById("registerBtn").addEventListener("click", () => {
    navigateTo("/register");
  });

  document.getElementById("guestBtn").addEventListener("click", () => {
    navigateTo("/guest/feed");
  });
}
