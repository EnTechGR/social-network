document.getElementById("logout-link").addEventListener("click", async (e) => {
  e.preventDefault();

  await fetch("http://localhost:8080/forum/api/session/logout", {
    method: "POST",
    credentials: "include", // Ensure cookies are sent
  });
  // Remove CSRF cookie and any stored token on the client
  // document.cookie = "csrf_token_frontend=; path=/; max-age=0";
  // localStorage.removeItem("csrfToken");
  // window.location.href = "/login";
});
