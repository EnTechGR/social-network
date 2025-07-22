document.getElementById("loginForm").addEventListener("submit", async (e) => {
  e.preventDefault();

  const emailInput = document.getElementById("email"); // Get the email input element
  const passwordInput = document.getElementById("password"); // Get the password input element

  const email = emailInput.value.trim();
  const password = passwordInput.value;
  const message = document.getElementById("message");

  try {
    const response = await fetch("http://localhost:8080/forum/api/session/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include", // IMPORTANT to send and receive cookies
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();

    if (response.ok) {
      message.textContent = data.message;
      message.style.color = "green";

      // Clear the form fields after successful submission
      emailInput.value = "";
      passwordInput.value = "";

      // Redirect after successful login
      setTimeout(() => {
        window.location.href = "/user/feed"; // Redirect to user page
      }, 1000);
    } else {
      message.textContent = data.message || "Login failed!";
      message.style.color = "red";
    }
  } catch (error) {
    message.textContent = "Error connecting to server.";
    message.style.color = "red";
  }
});

document.getElementById("googleRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/google/login";
});

document.getElementById("githubRegisterBtn").addEventListener("click", () => {
  window.location.href = "http://localhost:8080/auth/github/login";
});