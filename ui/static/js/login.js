document.getElementById("loginForm").addEventListener("submit", async (e) => {
  e.preventDefault();

  const emailInput = document.getElementById("email"); // This can be username or email now
  const passwordInput = document.getElementById("password");

  const login = emailInput.value.trim(); // Changed from 'email' to 'login'
  const password = passwordInput.value;
  const message = document.getElementById("message");

  try {
    const response = await fetch("http://localhost:8080/forum/api/session/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include", // IMPORTANT to send and receive cookies
      body: JSON.stringify({ login, password }), // Changed to use 'login' instead of 'email'
    });

    const data = await response.json();

    if (response.ok) {
      message.textContent = "Login successful!";
      message.style.color = "green";

      // Clear the form fields after successful submission
      emailInput.value = "";
      passwordInput.value = "";

      // Redirect after successful login
      setTimeout(() => {
        window.location.href = "/user/feed"; // Redirect to user page
      }, 1000);
    } else {
      message.textContent = data.error || "Login failed!";
      message.style.color = "red";
    }
  } catch (error) {
    message.textContent = "Error connecting to server.";
    message.style.color = "red";
    console.error("Login error:", error);
  }
});

document.getElementById("googleRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/google/login";
});

document.getElementById("githubRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/github/login";
});