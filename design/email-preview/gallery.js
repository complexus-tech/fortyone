"use strict";

const collection = JSON.parse(document.getElementById("manifest").textContent);
const frame = document.getElementById("email-frame");
const windowElement = document.getElementById("email-window");
const navigation = document.getElementById("email-navigation");
const metadata = document.getElementById("metadata");
const toggle = document.getElementById("metadata-toggle");
const dialog = document.getElementById("link-dialog");
let activeEmail;
let observer;

for (const group of ["Workspace", "Notifications", "Maya"]) {
  const heading = document.createElement("h3");
  heading.textContent = group;
  navigation.append(heading);
  for (const email of collection.filter((item) => item.group === group)) {
    const link = document.createElement("a");
    link.href = `#${email.id}`;
    link.dataset.id = email.id;
    const name = document.createElement("span");
    name.textContent = email.name;
    const number = document.createElement("span");
    number.className = "item-number";
    number.textContent = String(collection.indexOf(email) + 1).padStart(2, "0");
    link.append(name, number);
    navigation.append(link);
  }
}

function resizeFrame() {
  if (!frame.contentDocument?.body) return;
  // Body height, rather than root scrollHeight, permits shrinking after a switch.
  frame.style.height = `${Math.ceil(frame.contentDocument.body.getBoundingClientRect().height)}px`;
}

function applyImageSetting() {
  const emailDocument = frame.contentDocument;
  if (!emailDocument?.body) return;
  let style = emailDocument.getElementById("preview-image-toggle");
  if (!style) {
    style = emailDocument.createElement("style");
    style.id = "preview-image-toggle";
    emailDocument.head.append(style);
  }
  style.textContent = document.getElementById("hide-images").checked
    ? "img{display:none!important;}"
    : "";
  resizeFrame();
}

frame.addEventListener("load", () => {
  const emailDocument = frame.contentDocument;
  if (!emailDocument?.body) return;
  observer?.disconnect();
  observer = new ResizeObserver(resizeFrame);
  observer.observe(emailDocument.body);
  for (const image of emailDocument.images)
    image.addEventListener("load", resizeFrame, { once: true });
  emailDocument.addEventListener("click", (event) => {
    const link = event.target.closest("a");
    if (!link) return;
    event.preventDefault();
    document.getElementById("link-destination").textContent = link.href;
    dialog.showModal();
  });
  applyImageSetting();
  resizeFrame();
});

function selectEmail() {
  const email =
    collection.find((item) => item.id === location.hash.slice(1)) ||
    collection[0];
  if (email.id === activeEmail?.id) return;
  activeEmail = email;
  document.title = `${email.name} — FortyOne email studio`;
  document.getElementById("template-name").textContent = email.name;
  document.getElementById("subject").textContent = email.subject;
  document.getElementById("envelope-sender").textContent =
    `${email.sender} → ${email.recipient}`;
  document.getElementById("preheader").textContent = email.preheader;
  for (const link of navigation.querySelectorAll("a")) {
    if (link.dataset.id === email.id) link.setAttribute("aria-current", "page");
    else link.removeAttribute("aria-current");
  }
  for (const id of ["open-html", "download-html"])
    document.getElementById(id).href = email.html;
  document.getElementById("open-text").href = email.text;
  const details = document.getElementById("metadata-list");
  details.replaceChildren();
  for (const [name, value] of Object.entries({
    From: email.sender,
    To: email.recipient,
    Subject: email.subject,
    Preheader: email.preheader,
    "Reply-To": email.replyTo,
    "HTML size": `${(email.bytes / 1024).toFixed(1)} KB`,
  })) {
    const term = document.createElement("dt");
    const definition = document.createElement("dd");
    term.textContent = name;
    definition.textContent = value;
    details.append(term, definition);
  }
  document.getElementById("fields").textContent = JSON.stringify(
    email.fields,
    null,
    2,
  );
  document.getElementById("source-path").textContent = email.source;
  document.getElementById("template-note").textContent =
    email.note ||
    "Prototype sample values. The integrated gallery shows the application templates.";
  frame.title = `${email.name} email preview`;
  frame.src = email.html;
}

toggle.addEventListener("click", () => {
  metadata.hidden = !metadata.hidden;
  toggle.setAttribute("aria-expanded", String(!metadata.hidden));
});
for (const button of document.querySelectorAll("[data-width]")) {
  button.addEventListener("click", () => {
    windowElement.dataset.size = button.dataset.width;
    for (const sibling of document.querySelectorAll("[data-width]"))
      sibling.setAttribute("aria-pressed", String(sibling === button));
    resizeFrame();
  });
}
document
  .getElementById("hide-images")
  .addEventListener("change", applyImageSetting);
window.addEventListener("hashchange", selectEmail);
window.addEventListener("resize", resizeFrame);
selectEmail();
