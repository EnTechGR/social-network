import { navigateTo } from "../router.js";
import { getCSRF } from "../session.js";

export function renderLogin(app) {
  app.innerHTML = `
    <section class="login-page fade-in">
      <div class="login-card">
        <h1>Welcome Back</h1>
        <p class="subtitle">Sign in to continue your BookTalk journey.</p>

        <form id="loginForm" class="login-form">
          <input type="email" id="email" placeholder="Email" required />
          <input type="password" id="password" placeholder="Password" required />
          <button type="submit">Login</button>
          <p id="message" class="message"></p>
        </form>

        <div class="oauth-buttons">
          <button id="googleLoginBtn">Login with Google</button>
          <button id="githubLoginBtn">Login with GitHub</button>
        </div>

        <div class="login-footer">
          <p>Don’t have an account? <a href="#" id="registerLink">Register</a></p>
        </div>
      </div>
    </section>
  `;

  const form = document.getElementById("loginForm");
  const message = document.getElementById("message");

  const showMessage = (text, color) => {
    message.textContent = text;
    message.style.color = color;
  };

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    try {
      const res = await fetch("http://localhost:8080/forum/api/session/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(getCSRF() && { "X-CSRF-Token": getCSRF() }),
        },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        showMessage(data.message || "Login successful!", "green");
        form.reset();
        setTimeout(() => navigateTo("/user/feed"), 500);
      } else {
        showMessage(data.message || "Login failed!", "red");
      }
    } catch (err) {
      console.error(err);
      showMessage("Error connecting to server.", "red");
    }
  });

  document.getElementById("googleLoginBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/google/login";
  });

  document.getElementById("githubLoginBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/github/login";
  });

  document.getElementById("registerLink").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/register");
  });
}
