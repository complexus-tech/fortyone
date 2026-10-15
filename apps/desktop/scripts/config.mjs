import { fileURLToPath } from "node:url";

export const DEFAULT_DESKTOP_URL = "http://localhost:3000";

const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);
const NATIVE_CONFIG = fileURLToPath(
  new URL("../src-tauri/tauri.native.conf.json", import.meta.url),
);
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
  "TAURI_SIGNING_PRIVATE_KEY",
  "TAURI_SIGNING_PRIVATE_KEY_PASSWORD",
  "FORTYONE_UPDATER_PUBLIC_KEY",
  "FORTYONE_UPDATER_URL",
  "FORTYONE_WINDOWS_CERTIFICATE_THUMBPRINT",
  "FORTYONE_WINDOWS_TIMESTAMP_URL",
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

export function getNativeArguments(command, environment = {}) {
  const common = [
    "--features",
    "native-client",
    "--config",
    NATIVE_CONFIG,
    "--config",
    JSON.stringify(getFrontendSecurity(environment, false)),
  ];
  if (command === "dev") return ["dev", ...common];
  if (command === "build")
    return ["build", "--debug", "--no-bundle", ...common];
  throw new Error(`Unsupported native desktop command: ${command}`);
}

export function getProductionPackageArguments(
  platform = process.platform,
  releaseConfig,
) {
  if (!["darwin", "win32"].includes(platform)) {
    throw new Error("Desktop installers must be built on macOS or Windows.");
  }
  return [
    "build",
    "--features",
    "production",
    "--config",
    PRODUCTION_CONFIG,
    ...(releaseConfig ? ["--config", JSON.stringify(releaseConfig)] : []),
    "--bundles",
    platform === "darwin" ? "dmg" : "nsis",
    "--ci",
    "--",
    "--locked",
  ];
}

export function getProductionEnvironment(environment) {
  const env = { ...environment, APPLE_SIGNING_IDENTITY: "-" };
  // Local installers deliberately carry no distribution or updater credentials.
  for (const name of PRODUCTION_EXCLUDED_ENVIRONMENT) delete env[name];
  return env;
}

function requireValue(environment, name) {
  const value = environment[name]?.trim();
  if (!value) throw new Error(`Release packaging requires ${name}.`);
  return value;
}

export function getReleaseConfiguration(
  environment,
  platform = process.platform,
) {
  const publicKey = requireValue(environment, "FORTYONE_UPDATER_PUBLIC_KEY");
  const endpoint = requireValue(environment, "FORTYONE_UPDATER_URL");
  requireValue(environment, "TAURI_SIGNING_PRIVATE_KEY");
  let url;
  try {
    url = new URL(endpoint);
  } catch {
    throw new Error("FORTYONE_UPDATER_URL must be an absolute HTTPS URL.");
  }
  if (url.protocol !== "https:" || url.username || url.password || url.hash) {
    throw new Error(
      "FORTYONE_UPDATER_URL must use HTTPS without credentials or a fragment.",
    );
  }
  const config = {
    bundle: { createUpdaterArtifacts: true },
    plugins: { updater: { pubkey: publicKey, endpoints: [endpoint] } },
  };
  if (platform === "darwin") {
    const identity = requireValue(environment, "APPLE_SIGNING_IDENTITY");
    if (identity === "-")
      throw new Error(
        "Release packaging requires a Developer ID signing identity, not ad-hoc signing.",
      );
    const apiCredentials = [
      "APPLE_API_ISSUER",
      "APPLE_API_KEY",
      "APPLE_API_KEY_PATH",
    ].every((name) => environment[name]?.trim());
    const accountCredentials = [
      "APPLE_ID",
      "APPLE_PASSWORD",
      "APPLE_TEAM_ID",
    ].every((name) => environment[name]?.trim());
    if (!apiCredentials && !accountCredentials)
      throw new Error(
        "Release packaging requires complete Apple notarization credentials: APPLE_API_ISSUER/APPLE_API_KEY/APPLE_API_KEY_PATH or APPLE_ID/APPLE_PASSWORD/APPLE_TEAM_ID.",
      );
    config.bundle.macOS = { signingIdentity: identity };
  } else if (platform === "win32") {
    const thumbprint = requireValue(
      environment,
      "FORTYONE_WINDOWS_CERTIFICATE_THUMBPRINT",
    );
    if (!/^[a-fA-F0-9]{40}$/.test(thumbprint))
      throw new Error(
        "FORTYONE_WINDOWS_CERTIFICATE_THUMBPRINT must be a 40-character certificate thumbprint from the Windows certificate store.",
      );
    const timestampUrl = requireValue(
      environment,
      "FORTYONE_WINDOWS_TIMESTAMP_URL",
    );
    let timestamp;
    try {
      timestamp = new URL(timestampUrl);
    } catch {
      throw new Error(
        "FORTYONE_WINDOWS_TIMESTAMP_URL must be an absolute HTTP or HTTPS URL.",
      );
    }
    if (
      !["http:", "https:"].includes(timestamp.protocol) ||
      timestamp.username ||
      timestamp.password
    )
      throw new Error(
        "FORTYONE_WINDOWS_TIMESTAMP_URL must use HTTP or HTTPS without credentials.",
      );
    config.bundle.windows = {
      certificateThumbprint: thumbprint,
      digestAlgorithm: "sha256",
      timestampUrl,
    };
  } else {
    throw new Error(
      "Desktop release installers must be built on macOS or Windows.",
    );
  }
  return config;
}

export function getReleaseEnvironment(environment) {
  const env = { ...environment };
  delete env.FORTYONE_DESKTOP_URL;
  delete env.TAURI_CONFIG;
  return env;
}

export function getFrontendSecurity(environment, production = true) {
  const configured = environment.FORTYONE_COLLABORATION_URL?.trim();
  let origin = "";
  if (configured) {
    let url;
    try {
      url = new URL(configured);
    } catch {
      throw new Error(
        "FORTYONE_COLLABORATION_URL must be an absolute WebSocket URL.",
      );
    }
    const local =
      !production && url.protocol === "ws:" && LOOPBACK_HOSTS.has(url.hostname);
    if (
      (url.protocol !== "wss:" && !local) ||
      url.username ||
      url.password ||
      url.search ||
      url.hash
    )
      throw new Error(
        "FORTYONE_COLLABORATION_URL requires WSS (or local WS in development), without credentials, query parameters or a fragment.",
      );
    origin = url.origin;
  }
  const directives = `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; media-src 'self' blob:; connect-src ipc: http://ipc.localhost${origin ? ` ${origin}` : ""}; font-src 'self' data:; object-src 'none'; base-uri 'none'; frame-src 'none'`;
  return {
    app: {
      security: {
        csp: directives,
        ...(!production
          ? {
              devCsp: directives.replace(
                "connect-src ipc:",
                "connect-src http://localhost:5178 http://127.0.0.1:5178 ws://localhost:5178 ws://127.0.0.1:5178 ipc:",
              ),
            }
          : {}),
      },
    },
  };
}
