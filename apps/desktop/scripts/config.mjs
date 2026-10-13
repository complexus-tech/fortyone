import { fileURLToPath } from "node:url";

export const DEFAULT_DESKTOP_URL = "http://localhost:3000";

const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);
const PRODUCTION_CONFIG = fileURLToPath(
  new URL("../src-tauri/tauri.production.conf.json", import.meta.url),
);
const PRODUCTION_EXCLUDED_ENVIRONMENT = [
  "FORTYONE_DESKTOP_URL",
  "TAURI_CONFIG",
  "APPLE_CERTIFICATE",
  "APPLE_CERTIFICATE_PASSWORD",
  "APPLE_ID",
  "APPLE_PASSWORD",
  "APPLE_TEAM_ID",
  "APPLE_API_ISSUER",
  "APPLE_API_KEY",
  "APPLE_API_KEY_PATH",
];

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

export function getProductionPackageArguments(platform = process.platform) {
  if (platform !== "darwin") {
    throw new Error(
      "Run pnpm package:desktop on a Mac to create the macOS DMG.",
    );
  }

  return [
    "build",
    "--features",
    "production",
    "--config",
    PRODUCTION_CONFIG,
    "--bundles",
    "dmg",
    "--ci",
    "--",
    "--locked",
  ];
}

export function getProductionEnvironment(environment) {
  const env = { ...environment, APPLE_SIGNING_IDENTITY: "-" };

  // This local packaging path uses ad-hoc signing without Apple notarization.
  for (const name of PRODUCTION_EXCLUDED_ENVIRONMENT) delete env[name];

  return env;
}
