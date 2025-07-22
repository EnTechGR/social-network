document.getElementById("registerForm").addEventListener("submit", async (e) => {
  e.preventDefault();

  // Get references to all input elements
  const usernameInput = document.getElementById("username");
  const emailInput = document.getElementById("email");
  const passwordInput = document.getElementById("password");
  const confirmPasswordInput = document.getElementById("confirmPassword");

  const username = usernameInput.value.trim();
  const email = emailInput.value.trim();
  const password = passwordInput.value;
  const confirmPassword = confirmPasswordInput.value;
  const message = document.getElementById("message");

  if (password !== confirmPassword) {
    message.textContent = "Passwords do not match!";
    message.classList.remove("success");
    return;
  }

  try {
    const response = await fetch("http://localhost:8080/forum/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include", // ✅ Include cookies!
      body: JSON.stringify({ username, email, password }),
    });

    const data = await response.json();

    if (response.ok) {
      // ✅ Save CSRF token if needed
      if (data.csrf_token) {
        localStorage.setItem("csrfToken", data.csrf_token);
      }

      message.textContent = "Registration successful!";
      message.classList.add("success");

      // Clear the form fields after successful registration
      usernameInput.value = "";
      emailInput.value = "";
      passwordInput.value = "";
      confirmPasswordInput.value = "";

      setTimeout(() => {
        window.location.replace("/user/feed");
      }, 1000);
    } else {
      message.textContent = data.message || "Registration failed!";
      message.classList.remove("success");
    }
  } catch (error) {
    message.textContent = "Error connecting to server.";
    message.classList.remove("success");
  }
});

document.getElementById("googleRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/google/login";
});

document.getElementById("githubRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/github/login";
});