import { isStoryUuid } from "@/shared/routing/story";

type NativeLinkBase = { href: string; workspaceSlug: string };
export type NativeLinkTarget = NativeLinkBase &
  (
    | { kind: "story"; id: string }
    | { kind: "objective"; id: string; teamId: string }
    | { kind: "document"; id: string }
    | { kind: "unsupported" }
  );

const WORKSPACE_SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const STORY_REFERENCE = /^[a-z][a-z0-9]*-[1-9]\d*$/i;
const LOCAL_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);
const PUBLIC_HOSTS = new Set([
  "fortyone.app",
  "www.fortyone.app",
  "docs.fortyone.app",
  "status.fortyone.app",
]);
const NON_WORKSPACE_HOSTS = new Set(["api"]);

/** Null means external. Unsupported internal links must never be sent to OG services. */
export const parseNativeLink = (
  value: string,
  { workspaceSlug, origin }: { workspaceSlug: string; origin?: string },
): NativeLinkTarget | null => {
  if (!value.startsWith("/") && !/^https?:\/\//i.test(value)) return null;
  let url: URL;
  try {
    url = new URL(value, origin);
  } catch {
    return null;
  }
  if (
    !["http:", "https:"].includes(url.protocol) ||
    url.username ||
    url.password
  )
    return null;

  const host = url.hostname.toLowerCase();
  if (PUBLIC_HOSTS.has(host)) return null;
  const official = host === "fortyone.app" || host.endsWith(".fortyone.app");
  const local =
    LOCAL_HOSTS.has(host) && origin !== undefined && url.origin === origin;
  if (!official && !local) return null;

  // Keep even malformed/private app URLs away from external metadata fetches.
  const unsupported: NativeLinkTarget = {
    kind: "unsupported",
    href: url.href,
    workspaceSlug: "",
  };
  if (official && (url.protocol !== "https:" || url.port)) return unsupported;
  if (/%(?:2f|5c|00|25)/i.test(url.pathname) || value.includes("\\"))
    return unsupported;
  let parts: string[];
  try {
    parts = url.pathname.split("/").filter(Boolean).map(decodeURIComponent);
  } catch {
    return unsupported;
  }

  let targetWorkspace: string;
  if (
    value.startsWith("/") &&
    !value.startsWith("//") &&
    ["work", "teams", "docs"].includes(parts[0])
  ) {
    targetWorkspace = workspaceSlug.toLowerCase();
  } else if (host === "cloud.fortyone.app") {
    targetWorkspace = parts.shift()?.toLowerCase() ?? "";
  } else if (official) {
    const subdomain = host.slice(0, -".fortyone.app".length);
    if (!WORKSPACE_SLUG.test(subdomain) || NON_WORKSPACE_HOSTS.has(subdomain))
      return unsupported;
    targetWorkspace = subdomain;
  } else {
    const first = parts[0];
    targetWorkspace = ["work", "teams", "docs"].includes(first)
      ? workspaceSlug.toLowerCase()
      : parts.shift()?.toLowerCase() ?? "";
  }
  if (!WORKSPACE_SLUG.test(targetWorkspace)) return unsupported;
  const base = { href: url.href, workspaceSlug: targetWorkspace };
  if (
    parts.length === 2 &&
    parts[0] === "work" &&
    (isStoryUuid(parts[1]) || STORY_REFERENCE.test(parts[1]))
  ) {
    return {
      ...base,
      kind: "story",
      id: isStoryUuid(parts[1])
        ? parts[1].toLowerCase()
        : parts[1].toUpperCase(),
    };
  }
  if (
    parts.length === 4 &&
    parts[0] === "teams" &&
    isStoryUuid(parts[1]) &&
    parts[2] === "objectives" &&
    isStoryUuid(parts[3])
  ) {
    return {
      ...base,
      kind: "objective",
      teamId: parts[1].toLowerCase(),
      id: parts[3].toLowerCase(),
    };
  }
  if (parts.length === 2 && parts[0] === "docs" && isStoryUuid(parts[1])) {
    return { ...base, kind: "document", id: parts[1].toLowerCase() };
  }
  return { ...base, kind: "unsupported" };
};
