import { navigateTo } from "../router.js";
import { getCSRF, verifySession } from "../session.js";

export function renderRegister(app) {
  app.innerHTML = `
    <section class="register-page fade-in">
      <div class="register-card">
        <h1>Join BookTalk</h1>
        <p class="subtitle">Create an account to start sharing your thoughts.</p>

        <form id="registerForm" class="register-form">
          <input type="text" id="username" placeholder="Username" required />
          <input type="email" id="email" placeholder="Email" required />

          <input type="text" id="firstName" placeholder="First name" required />
          <input type="text" id="lastName" placeholder="Last name" required />

          <input
            type="number"
            id="age"
            placeholder="Age"
            min="13"
            max="120"
            required
          />

          <select id="gender" required>
            <option value="">Select gender</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="other">Other</option>
            <option value="prefer_not_to_say">Prefer not to say</option>
          </select>

          <input
            type="password"
            id="password"
            placeholder="Password"
            required
          />
          <input
            type="password"
            id="confirmPassword"
            placeholder="Confirm Password"
            required
          />

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
    const firstName = document.getElementById("firstName").value.trim();
    const lastName = document.getElementById("lastName").value.trim();
    const ageValue = document.getElementById("age").value.trim();
    const gender = document.getElementById("gender").value;
    const password = document.getElementById("password").value;
    const confirmPassword = document.getElementById("confirmPassword").value;

    if (password !== confirmPassword) {
      showMessage("Passwords do not match!");
      return;
    }

    const age = parseInt(ageValue, 10);
    if (Number.isNaN(age) || age < 13 || age > 120) {
      showMessage("Age must be between 13 and 120");
      return;
    }

    const allowedGenders = ["male", "female", "other", "prefer_not_to_say"];
    if (!allowedGenders.includes(gender)) {
      showMessage(
        "Gender must be one of: male, female, other, prefer_not_to_say"
      );
      return;
    }

    if (!username || !email || !firstName || !lastName || !gender) {
      showMessage(
        "All fields are required: username, email, password, first_name, last_name, age, gender"
      );
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
        body: JSON.stringify({
          username,
          email,
          password,
          first_name: firstName,
          last_name: lastName,
          age,
          gender,
        }),
      });

      let data = {};
      try {
        data = await res.json();
      } catch {
        // ignore JSON parse errors
      }

      if (res.ok) {
        // ✅ Ensure SPA state matches backend session *after* registration
        await verifySession();

        showMessage("Registration successful!", true);
        form.reset();
        setTimeout(() => navigateTo("/user/feed"), 500);
      } else {
        const errorMsg =
          data.error ||
          data.message ||
          "Registration failed! Please check your details and try again.";
        showMessage(errorMsg);
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
