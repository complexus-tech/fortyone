import { cp, mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { join } from "node:path";

const root = fileURLToPath(new URL(".", import.meta.url));
const output = join(root, "integrated");
const assets = fileURLToPath(
  new URL("../../apps/landing/public/email-assets/v1/", import.meta.url),
);
await mkdir(output, { recursive: true });
await cp(assets, join(output, "assets"), { recursive: true });
for (const file of ["gallery.css", "gallery.js"])
  await cp(join(root, file), join(output, file));
const names = {
  "invites-invitation": "Workspace invitation",
  "invites-acceptance": "Invitation accepted",
  "notifications-notification": "Story notification",
  "maya-weekly": "Maya’s weekly note",
  "maya-confirmation": "Maya’s confirmation",
  "auth-verification": "Login link",
  "auth-verification_mobile": "Login code",
  "feedback-verification": "Feedback verification",
  "users-inactivity_warning": "Inactive account",
  "workspaces-inactivity_warning": "Inactive workspace",
  "workspaces-deletion_scheduled_confirmation": "Deletion confirmation",
  "workspaces-deletion_scheduled_notification": "Deletion notification",
  "workspaces-restored_confirmation": "Restoration confirmation",
  "workspaces-restored_notification": "Restoration notification",
};
const files = await readdir(join(root, "rendered"));
const manifest = [];
for (const [id, name] of Object.entries(names)) {
  const file = `${id}.html`;
  if (!files.includes(file)) continue;
  const original = await readFile(join(root, "rendered", file), "utf8");
  // Only asset locations change for local inspection; markup is Go-rendered.
  const html = original.replaceAll(
    "https://fortyone.app/email-assets/v1/",
    "assets/",
  );
  await writeFile(join(output, file), html);
  const maya = id.startsWith("maya-");
  manifest.push({
    id,
    name,
    group: maya
      ? "Maya"
      : id.startsWith("invites-") || id.startsWith("workspaces-")
        ? "Workspace"
        : "Notifications",
    subject: name,
    preheader: "Rendered by the application mailer with sample data.",
    html: file,
    text: file,
    sender: maya
      ? "Maya <maya@example.com>"
      : "FortyOne <notifications@example.com>",
    recipient: "Sample recipient",
    replyTo: maya
      ? "Existing per-thread Reply-To is preserved"
      : "Existing sender settings",
    bytes: Buffer.byteLength(original),
    fields: {},
    source:
      id === "maya-confirmation"
        ? "apps/server/internal/modules/emailreply/service/reply_html.go"
        : id === "maya-weekly"
          ? "apps/server/templates/notifications/notification.html"
          : `apps/server/templates/${id.replace("-", "/")}.html`,
    note: "Application template preview. Only image/font URLs are localized. Destination links are intercepted in this gallery.",
  });
}
let gallery = await readFile(join(root, "gallery.html"), "utf8");
gallery = gallery
  .replace(
    "<!-- MANIFEST -->",
    `<script id="manifest" type="application/json">${JSON.stringify(manifest).replaceAll("<", "\\u003c")}</script>`,
  )
  .replace("EMAIL DESIGN STUDIO", "APPLICATION EMAILS")
  .replace("Preview collection", "Go-rendered templates")
  .replace('href="email-preview.zip"', 'href="../email-preview.zip"')
  .replace(
    "Preview only. Email links use example.com.",
    "Application templates · Sample data · Links intercepted.",
  )
  .replace(
    '<a id="open-text" target="_blank" rel="noopener">Plain text</a>',
    '<a id="open-text" target="_blank" rel="noopener" hidden>HTML</a>',
  );
await writeFile(join(output, "index.html"), gallery);
console.log(`Built ${manifest.length} previews from the Go mailer output.`);
