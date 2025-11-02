import { navigateTo } from "../router.js";
import { getCSRF } from "../session.js";

export function renderLogin(app) {
  app.innerHTML = `
    <section class="login-page">
      <div class="login-card">
        <h1>Login</h1>

        <form id="loginForm" class="login-form">
          <input type="email" id="email" placeholder="Email" required />
          <input type="password" id="password" placeholder="Password" required />
          <button type="submit">Login</button>
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
          "X-CSRF-Token": getCSRF(), // Include CSRF token for security
        },
        credentials: "include",
        body: JSON.stringify({ email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        showMessage(data.message || "Login successful!", "green");

        form.reset(); // Clear fields

        // SPA navigation instead of full page reload
        setTimeout(() => {
          navigateTo("/user/feed");
        }, 500);
      } else {
        showMessage(data.message || "Login failed!", "red");
      }
    } catch (err) {
      console.error(err);
      showMessage("Error connecting to server.", "red");
    }
  });

  // OAuth login buttons
  document.getElementById("googleRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/google/login";
  });

  document.getElementById("githubRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/github/login";
  });

  // SPA navigation to register page
  document.getElementById("toRegister").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/register");
  });
}
