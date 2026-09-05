import { readFile } from "node:fs/promises";
import { stripTypeScriptTypes } from "node:module";

// Execute the frontend source so preview colors cannot drift from the app.
const source = await readFile(
  new URL("../../packages/lib/src/avatar-color.ts", import.meta.url),
  "utf8",
);
const moduleURL = `data:text/javascript;base64,${Buffer.from(stripTypeScriptTypes(source)).toString("base64")}`;
export const { getAvatarColor } = await import(moduleURL);

// Match packages/ui/src/avatar.tsx, with a safe fallback for whitespace-only names.
export function getInitials(name) {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (!words.length) return "U";
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words.at(-1)[0]).toUpperCase();
}

export function avatarImageURL(value) {
  if (typeof value !== "string" || !value.trim()) return null;
  // Relative paths are limited to packaged sample portraits. Real photos use HTTPS.
  if (/^\.\.\/assets\/avatars\/[a-z0-9-]+\.(png|jpe?g|webp)$/i.test(value))
    return value;
  try {
    const url = new URL(value);
    return url.protocol === "https:" && !url.username && !url.password
      ? url.href
      : null;
  } catch {
    return null;
  }
}
