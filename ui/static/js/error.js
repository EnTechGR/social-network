// /static/js/error.js

window.addEventListener('DOMContentLoaded', () => {
  const params = new URLSearchParams(window.location.search);
  const msg = params.get('msg');
  const code = params.get('code');

  const errorMessage = document.getElementById('error-message');
  const errorCode = document.getElementById('error-code');

  if (errorCode) {
    errorCode.textContent = code ? `Error ${code}` : 'Error';
  }

  if (errorMessage) {
    errorMessage.textContent = msg ? decodeURIComponent(msg) : "An unexpected error occurred.";
  }
  document.getElementById("back-button")?.addEventListener("click", () => {
  window.history.back();
});

});
