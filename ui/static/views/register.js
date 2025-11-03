import { navigateTo } from "../router.js";
import { getCSRF } from "../session.js";

export function renderRegister(app) {
  app.innerHTML = `
    <section class="register-page fade-in">
      <div class="register-card">
        <h1>Join BookTalk</h1>
        <p class="subtitle">Create an account to start sharing your thoughts.</p>

        <form id="registerForm" class="register-form">
          <input type="text" id="username" placeholder="Username" required />
          <input type="email" id="email" placeholder="Email" required />
          <input type="password" id="password" placeholder="Password" required />
          <input type="password" id="confirmPassword" placeholder="Confirm Password" required />
          <button type="submit">Register</button>
          <p id="message" class="message"></p>
        </form>

        <div class="oauth-buttons">
          <button id="googleRegisterBtn">Register with Google</button>
          <button id="githubRegisterBtn">Register with GitHub</button>
        </div>

        <div class="register-footer">
          <p>Already have an account? <a href="#" id="loginLink">Login</a></p>
        </div>
      </div>
    </section>
  `;

  const form = document.getElementById("registerForm");
  const message = document.getElementById("message");

  const showMessage = (text, success = false) => {
    message.textContent = text;
    message.style.color = success ? "green" : "red";
  };

  form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const username = document.getElementById("username").value.trim();
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;
    const confirmPassword = document.getElementById("confirmPassword").value;

    if (password !== confirmPassword) {
      showMessage("Passwords do not match!");
      return;
    }

    try {
      const res = await fetch("http://localhost:8080/forum/api/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...(getCSRF() && { "X-CSRF-Token": getCSRF() }),
        },
        credentials: "include",
        body: JSON.stringify({ username, email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        showMessage("Registration successful!", true);
        form.reset();
        setTimeout(() => navigateTo("/user/feed"), 500);
      } else {
        showMessage(data.message || "Registration failed!");
      }
    } catch (err) {
      console.error(err);
      showMessage("Error connecting to server.");
    }
  });

  document.getElementById("googleRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/google/login";
  });

  document.getElementById("githubRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/github/login";
  });

  document.getElementById("loginLink").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/login");
  });
}
