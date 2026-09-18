export function isAllowedOrigin(origin: string, allowed: string[]) {
  if (allowed.includes(origin)) return true;
  let url: URL;
  try {
    url = new URL(origin);
  } catch {
    return false;
  }
  if (url.origin !== origin || url.protocol !== "https:") return false;
  return allowed.some((entry) => {
    if (!entry.startsWith("https://*.")) return false;
    const suffix = entry.slice("https://*".length);
    return url.host.endsWith(suffix) && url.host.length > suffix.length;
  });
}
