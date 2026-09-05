import { cp, mkdir, readFile, writeFile, access } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { join, resolve } from "node:path";
import { createRequire } from "node:module";
import { render, plainText } from "./render.mjs";
import { reviewEmails } from "./review-content.mjs";

const root = fileURLToPath(new URL(".", import.meta.url));
const out = join(root, "copy-review");
const repository = resolve(root, "../..");
const require = createRequire(import.meta.url);
// Use the project dependency when present; SHARP_MODULE supports the desktop runtime.
const sharp = require(process.env.SHARP_MODULE || "sharp");
await mkdir(join(out, "emails"), { recursive: true });
await cp(join(root, "assets"), join(out, "assets"), { recursive: true });
await mkdir(join(out, "assets/icons"), { recursive: true });
const shapes = {
  check: '<path d="M5 14L8.5 17.5L19 6.5"/>',
  clock: '<circle cx="12" cy="12" r="10"/><path d="M12 8V12L14 14"/>',
  lock: '<rect x="5" y="10" width="14" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3M12 14v3"/>',
};
for (const [name, shape] of Object.entries(shapes)) {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96" viewBox="0 0 32 32"><rect width="32" height="32" rx="9" fill="#fff0e5"/><g transform="translate(7 7) scale(.75)" fill="none" stroke="#a9512a" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${shape}</g></svg>`;
  await writeFile(join(out, `assets/icons/${name}.svg`), svg);
  await sharp(Buffer.from(svg))
    .png()
    .toFile(join(out, `assets/icons/${name}.png`));
}
await writeFile(
  join(out, "assets/icons/README.md"),
  "Check and clock paths adapted from packages/icons/src/check.tsx and clock.tsx. Lock is a simple purpose-built vector. PNG files are rendered at 3x their 32px display size; SVG sources are retained for editing. Email markup uses PNG only. Icons are decorative; the adjacent text carries their meaning.\n",
);
const ids = new Set();
const manifest = [];
for (const email of reviewEmails) {
  if (ids.has(email.id)) throw new Error(`Duplicate preview: ${email.id}`);
  ids.add(email.id);
  await access(join(repository, email.source));
  const html = render(email);
  if (/undefined|\[object Object\]/.test(html))
    throw new Error(`Incomplete fixture: ${email.id}`);
  await writeFile(join(out, `emails/${email.id}.html`), html);
  await writeFile(join(out, `emails/${email.id}.txt`), plainText(email));
  manifest.push({
    ...email,
    html: `emails/${email.id}.html`,
    text: `emails/${email.id}.txt`,
    bytes: Buffer.byteLength(html),
    sender: email.maya
      ? "Maya, AI Agent <maya@example.com>"
      : "FortyOne <notifications@example.com>",
    recipient: "Alex Morgan <alex@example.com>",
    replyTo: email.maya
      ? "Maya <maya-thread@example.com> (sample)"
      : "Not enabled in this preview",
    fields: {
      ...email.fields,
      ...(email.person ? { actor: email.person } : {}),
      ...(email.code ? { sampleCode: email.code } : {}),
      ...(email.rows ? { sampleDetails: email.rows } : {}),
    },
  });
}
await writeFile(
  join(out, "metadata.json"),
  JSON.stringify(manifest, null, 2) + "\n",
);
const gallery = (await readFile(join(root, "gallery.html"), "utf8"))
  .replace("A warmer inbox.", "Words that work.")
  .replace("EMAIL DESIGN STUDIO", "EMAIL COPY REVIEW")
  .replace(
    "One family. A little personality.<br />The right message for the moment.",
    `${manifest.length} proposed emails and variants.<br />Fictional data. Ready for review.`,
  )
  .replace(
    '<nav id="email-navigation"',
    '<label class="search-label" for="email-search">Find an email</label><input id="email-search" type="search" placeholder="Search emails…" autocomplete="off"><p id="search-count" aria-live="polite"></p><nav id="email-navigation"',
  )
  .replace('href="email-preview.zip"', 'href="../email-preview.zip"')
  .replace(
    'Live HTML <span aria-hidden="true">·</span> Sample content',
    "Proposed copy · Production unchanged",
  )
  .replace(
    "Preview only. Email links use example.com.",
    '<a href="COPY.md" target="_blank">Read all proposed copy</a> · <a href="COVERAGE.md" target="_blank">Coverage</a>',
  )
  .replace(
    "<!-- MANIFEST -->",
    `<script id="manifest" type="application/json">${JSON.stringify(manifest).replaceAll("<", "\\u003c")}</script>`,
  );
await writeFile(join(out, "index.html"), gallery);
let script = (await readFile(join(root, "gallery.js"), "utf8")).replace(
  '["Workspace", "Notifications", "Maya"]',
  "[...new Set(collection.map((email) => email.group))]",
);
script += `
const search = document.getElementById("email-search");
function filterEmails() {
  const query = search.value.trim().toLowerCase();
  let count = 0;
  for (const link of navigation.querySelectorAll("a")) {
    const email = collection.find((item) => item.id === link.dataset.id);
    link.hidden = ![email.name, email.group, email.title, email.subject].join(" ").toLowerCase().includes(query);
    if (!link.hidden) count++;
  }
  for (const heading of navigation.querySelectorAll("h3")) {
    let next = heading.nextElementSibling;
    let visible = false;
    while (next && next.tagName !== "H3") {
      if (!next.hidden) visible = true;
      next = next.nextElementSibling;
    }
    heading.hidden = !visible;
  }
  document.getElementById("search-count").textContent = count + " of " + collection.length + " previews";
}
search.addEventListener("input", filterEmails);
filterEmails();
`;
await writeFile(join(out, "gallery.js"), script);
await writeFile(
  join(out, "gallery.css"),
  (await readFile(join(root, "gallery.css"), "utf8")) +
    `
.search-label{font-size:12px;font-weight:500;margin:12px 0 7px;}
#email-search{width:100%;min-height:36px;padding:8px 10px;border:1px solid var(--line);border-radius:6px;background:#fff;font:inherit;font-size:13px;color:#25150e;}
#search-count{font-size:11px;color:var(--muted);margin:8px 0 4px;}
#email-navigation [hidden]{display:none!important;}
`,
);
await writeFile(
  join(out, "COPY.md"),
  `# FortyOne email copy review\n\n${manifest.length} proposed specimens. All names, dates, metrics, message excerpts, codes, and links are fictional. Production copy remains unchanged.\n\n` +
    manifest
      .map(
        (email) =>
          `## ${email.name}\n\n- Subject: ${email.subject}\n- Preheader: ${email.preheader}\n- Source: \`${email.source}\`\n\n${plainText(email)}\n`,
      )
      .join("\n"),
);
await writeFile(
  join(out, "COVERAGE.md"),
  `# Coverage and implementation notes\n\nThis collection covers the 12 file-backed application template types plus notification, job, and Maya reply variants found in the repository. The shared notification template carries many event types; these are representative states, not separate delivery templates. Dynamic AI wording, arbitrary field combinations, and provider-side Brevo automations are not an exhaustive enumerable catalog. No direct-message chat email producer was found; the conversation preview represents comment replies.\n\n## Review boundary\n\nOnly prototype files changed. Existing integrated previews continue to represent application templates. Subjects, preheaders, bodies, CTA labels, and footers here are proposed together. All values and destinations are fictional. Production implementation must preserve permission filtering, actual event facts, recipient timezone, token expiry, and unsubscribe/reply behavior. Maya replies retain literal CONFIRM and CANCEL commands; confirmation proposes one mutation.\n\n## Icons and avatars\n\nPNG icons are optional decoration for access, time-sensitive notices, and success receipts. They remain comprehensible with images blocked. Initials are rendered as text, with a visible full name beside them. Real avatar photos can be supported later by passing a suitable avatar URL through the notification payload; that plumbing is not implemented here. Do not fabricate profile photos or imply these initials are real user data.\n\n## Rollout copy locations\n\nInvitation subjects exist in both the event consumer and active invitation worker. Notification copy comes from rules and task handlers; digests come from jobs; Maya content comes from the email agent and deterministic reply processor. Editing template HTML alone will not update every subject or dynamically produced message.\n\n## Preview map\n\n| Preview | Group | Source |\n|---|---|---|\n` +
    manifest
      .map(
        (e) =>
          `| [${e.name}](emails/${e.id}.html) | ${e.group} | ${e.source} |`,
      )
      .join("\n") +
    "\n",
);
console.log(`Built ${manifest.length} copy review previews in ${out}`);
