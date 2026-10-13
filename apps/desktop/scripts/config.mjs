export const DEFAULT_DESKTOP_URL = "http://localhost:3000";

const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

export function getDesktopUrl(value = DEFAULT_DESKTOP_URL) {
  let url;

  try {
    url = new URL(value);
  } catch {
    throw new Error(
      "FORTYONE_DESKTOP_URL must be an absolute local URL, such as http://localhost:3000.",
    );
  }

  if (
    !["http:", "https:"].includes(url.protocol) ||
    !LOOPBACK_HOSTS.has(url.hostname) ||
    url.username ||
    url.password ||
    url.href.includes("#")
  ) {
    throw new Error(
      "FORTYONE_DESKTOP_URL must use HTTP or HTTPS on localhost, 127.0.0.1, or [::1], without credentials or a fragment. This desktop build is for local development.",
    );
  }

  return url.href;
}

export function getTauriArguments(command, url) {
  const config = JSON.stringify({ build: { devUrl: getDesktopUrl(url) } });

  if (command === "dev") {
    return ["dev", "--config", config];
  }

  if (command === "build") {
    return ["build", "--debug", "--no-bundle", "--config", config];
  }

  throw new Error(`Unsupported desktop command: ${command}`);
}
