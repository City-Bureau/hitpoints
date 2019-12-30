(function() {
  if (!window.hitpoints_tracked) {
    img = document.createElement("img");
    img.setAttribute("src", "%s://%s/pixel.gif?url=" + window.location.href);
    img.setAttribute("width", "1");
    img.setAttribute("height", "1");
    document.body.appendChild(img);
    window.hitpoints_tracked = true;
  }
})();
