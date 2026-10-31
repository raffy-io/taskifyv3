// HTMX related
document.addEventListener("DOMContentLoaded", () => {
  document.body.addEventListener("htmx:beforeSwap", (evt) => {
    // Add any status codes you want to allow to swap here
    const allowedStatuses = [400, 422, 409];

    if (allowedStatuses.includes(evt.detail.xhr.status)) {
      evt.detail.shouldSwap = true;
    }
  });
});

// Alpine related
document.addEventListener("alpine:init", () => {});
