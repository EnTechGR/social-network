export function renderNotFound(app) {
  app.innerHTML = `
    <div class="not-found">
      <h1>404</h1>
      <p>Page not found.</p>
      <button id="backHome">Go Home</button>
    </div>
  `;
  document
    .getElementById("backHome")
    .addEventListener("click", () => navigateTo("/"));
}
