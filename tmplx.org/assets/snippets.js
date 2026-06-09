// Fills .snippet-code[data-snippet] blocks with tree-sitter-highlighted
// fragments from /snippets/<name>.html, and wires up copy buttons.
(function () {
  document.querySelectorAll(".snippet-code[data-snippet]").forEach(function (el) {
    fetch("/snippets/" + el.getAttribute("data-snippet") + ".html")
      .then(function (r) { return r.text(); })
      .then(function (html) { el.innerHTML = html; })
      .catch(function () { el.textContent = "(failed to load snippet)"; });
  });
  document.querySelectorAll(".copy-btn").forEach(function (btn) {
    btn.addEventListener("click", function () {
      var code = btn.closest(".snippet").querySelector("code");
      if (!code) return;
      navigator.clipboard.writeText(code.textContent).then(function () {
        btn.textContent = "copied";
        setTimeout(function () { btn.textContent = "copy"; }, 1200);
      });
    });
  });
})();
