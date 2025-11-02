import { navigateTo } from "../router.js";
import { getCSRF } from "../session.js";

export function renderRegister(app) {
  app.innerHTML = `
    <h2>Register</h2>
    <form id="registerForm">
      <input type="text" id="username" placeholder="Username" required />
      <input type="email" id="email" placeholder="Email" required />
      <input type="password" id="password" placeholder="Password" required />
      <input type="password" id="confirmPassword" placeholder="Confirm Password" required />
      <button type="submit">Register</button>
    </form>
    <p id="message"></p>
    <button id="googleRegisterBtn">Register with Google</button>
    <button id="githubRegisterBtn">Register with GitHub</button>
    <p>Already have an account? <a href="/login" id="toLogin">Login</a></p>
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
          "X-CSRF-Token": getCSRF(),
        },
        credentials: "include",
        body: JSON.stringify({ username, email, password }),
      });

      const data = await res.json();

      if (res.ok) {
        showMessage("Registration successful!", true);
        form.reset();

        // SPA navigation
        setTimeout(() => {
          navigateTo("/user/feed");
        }, 500);
      } else {
        showMessage(data.message || "Registration failed!");
      }
    } catch (err) {
      console.error(err);
      showMessage("Error connecting to server.");
    }
  });

  // OAuth login buttons
  document.getElementById("googleRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/google/login";
  });

  document.getElementById("githubRegisterBtn").addEventListener("click", () => {
    window.location.href = "http://localhost:8080/auth/github/login";
  });

  // SPA navigation to login page
  document.getElementById("toLogin").addEventListener("click", (e) => {
    e.preventDefault();
    navigateTo("/login");
  });
}
