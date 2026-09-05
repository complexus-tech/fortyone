import { mkdir, readFile, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { join } from "node:path";
import { emails } from "./content.mjs";

const root = fileURLToPath(new URL(".", import.meta.url));
import { render, plainText } from "./render.mjs";

await mkdir(join(root, "emails"), { recursive: true });
const manifest = [];
for (const email of emails) {
  const html = render(email);
  await writeFile(join(root, "emails", `${email.id}.html`), html);
  await writeFile(join(root, "emails", `${email.id}.txt`), plainText(email));
  manifest.push({
    ...email,
    html: `emails/${email.id}.html`,
    text: `emails/${email.id}.txt`,
    bytes: Buffer.byteLength(html),
    sender: email.maya
      ? "Maya, AI Agent <maya@example.com>"
      : "FortyOne <notifications@example.com>",
    recipient: email.recipient || "Alex <alex@example.com>",
    replyTo: email.maya
      ? "Maya <maya-thread@example.com> (sample)"
      : "Not enabled in this preview",
  });
}
await writeFile(
  join(root, "metadata.json"),
  JSON.stringify(manifest, null, 2) + "\n",
);
const gallery = await readFile(join(root, "gallery.html"), "utf8");
await writeFile(
  join(root, "index.html"),
  gallery.replace(
    "<!-- MANIFEST -->",
    `<script id="manifest" type="application/json">${JSON.stringify(manifest).replaceAll("<", "\\u003c")}</script>`,
  ),
);
console.log(
  `Built ${manifest.length} HTML emails, plain-text alternatives, metadata, and gallery.`,
);
