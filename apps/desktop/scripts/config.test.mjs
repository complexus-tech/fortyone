import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import test from "node:test";
import { fileURLToPath } from "node:url";
import {
  getDesktopUrl,
  getFrontendSecurity,
  getNativeArguments,
  getReleaseConfiguration,
  getReleaseEnvironment,
  getProductionEnvironment,
  getProductionPackageArguments,
  getTauriArguments,
} from "./config.mjs";

test("local URLs preserve the requested port and app route", () => {
  assert.equal(getDesktopUrl(), "http://localhost:3000/");
  assert.equal(
    getDesktopUrl("https://localhost:3443/acme/stories?view=board"),
    "https://localhost:3443/acme/stories?view=board",
  );
  assert.equal(
    getDesktopUrl("http://127.0.0.1:3001"),
    "http://127.0.0.1:3001/",
  );
  assert.equal(getDesktopUrl("http://[::1]:3001"), "http://[::1]:3001/");
});

test("the local launcher rejects remote, credentialed, and non-web URLs", () => {
  for (const url of [
    "",
    "localhost:3000",
    "file:///etc/passwd",
    "javascript:alert(1)",
    "http://example.com",
    "https://example.com",
    "http://localhost.example.com:3000",
    "http://localhost@evil.example:3000",
    "http://user:password@localhost:3000",
    "http://localhost:3000/#token",
    "http://localhost:3000/#",
  ]) {
    assert.throws(() => getDesktopUrl(url), /FORTYONE_DESKTOP_URL/, url);
  }
});

test("local builds retain the custom URL and cannot become release installers", () => {
  const args = getTauriArguments("build", "http://localhost:3001/acme");
  assert.equal(args[0], "build");
  assert.ok(args.includes("--debug"));
  assert.ok(args.includes("--no-bundle"));
  assert.deepEqual(JSON.parse(args[args.indexOf("--config") + 1]), {
    build: { devUrl: "http://localhost:3001/acme" },
  });
  assert.throws(() => getTauriArguments("release", "http://localhost:3000"));
});

test("production packaging uses the release feature and a fixed DMG configuration", () => {
  const args = getProductionPackageArguments("darwin");
  assert.equal(args[0], "build");
  assert.equal(
    args[args.indexOf("--features") + 1],
    "production",
  );
  assert.equal(args[args.indexOf("--bundles") + 1], "dmg");
  assert.deepEqual(args.slice(-2), ["--", "--locked"]);
  for (const flag of ["--debug", "--no-bundle", "--no-sign"]) {
    assert.ok(!args.includes(flag), flag);
  }

  const config = JSON.parse(
    readFileSync(args[args.indexOf("--config") + 1], "utf8"),
  );
  assert.equal(config.productName, "FortyOne");
  assert.equal(config.identifier, "app.fortyone.desktop");
  assert.deepEqual(config.plugins["deep-link"].desktop.schemes, [
    "fortyone-desktop",
  ]);
  assert.deepEqual(config.build, {
    devUrl: null,
    frontendDist: "https://cloud.fortyone.app",
  });
  assert.equal(config.bundle.active, true);
  assert.deepEqual(config.bundle.targets, ["dmg"]);
  assert.equal(config.bundle.macOS.signingIdentity, "-");

  assert.throws(
    () => getProductionPackageArguments("linux"),
    /macOS or Windows/,
  );
  const windows = getProductionPackageArguments("win32");
  assert.equal(windows[windows.indexOf("--bundles") + 1], "nsis");
});

test("production packaging isolates local URLs, config overrides, and Apple credentials", () => {
  const environment = Object.freeze({
    PATH: "/usr/bin:/custom/cargo/bin",
    RUST_LOG: "info",
    FORTYONE_DESKTOP_URL: "file:///tmp/not-a-web-app.html",
    TAURI_CONFIG: '{"build":{"frontendDist":"http://localhost:39999"}}',
    APPLE_SIGNING_IDENTITY: "Developer ID Application: Test",
    APPLE_CERTIFICATE: "test-certificate",
    APPLE_CERTIFICATE_PASSWORD: "test-password",
    APPLE_ID: "test@example.com",
    APPLE_PASSWORD: "test-password",
    APPLE_TEAM_ID: "test-team",
    APPLE_API_ISSUER: "test-issuer",
    APPLE_API_KEY: "test-key",
    APPLE_API_KEY_PATH: "/tmp/test-key.p8",
  });

  assert.deepEqual(getProductionEnvironment(environment), {
    PATH: environment.PATH,
    RUST_LOG: "info",
    APPLE_SIGNING_IDENTITY: "-",
  });
  assert.equal(
    environment.FORTYONE_DESKTOP_URL,
    "file:///tmp/not-a-web-app.html",
  );
  assert.equal(
    environment.APPLE_SIGNING_IDENTITY,
    "Developer ID Application: Test",
  );
});

test("the packaging entry point rejects flags before invoking a build", () => {
  const launcher = fileURLToPath(new URL("./desktop.mjs", import.meta.url));
  const result = spawnSync(process.execPath, [launcher, "package", "--debug"], {
    encoding: "utf8",
  });
  assert.equal(result.status, 1);
  assert.match(result.stderr, /do not accept extra flags/);
  assert.match(result.stderr, /fixed production configuration/);
  assert.equal(result.stdout, "");
});

test("native development starts the bundled frontend with the isolated native feature", () => {
  const args = getNativeArguments("dev");
  assert.equal(args[0], "dev");
  assert.equal(args[args.indexOf("--features") + 1], "native-client");
  const config = JSON.parse(
    readFileSync(args[args.indexOf("--config") + 1], "utf8"),
  );
  assert.equal(config.build.frontendDist, "../dist");
  assert.equal(config.build.beforeDevCommand, "pnpm frontend:dev");
  assert.deepEqual(config.plugins["deep-link"].desktop.schemes, [
    "fortyone-desktop-dev",
  ]);
  assert.ok(getNativeArguments("build").includes("--no-bundle"));
});

const releaseEnvironment = {
  FORTYONE_UPDATER_PUBLIC_KEY: "public-signing-key",
  FORTYONE_UPDATER_URL:
    "https://updates.example.com/{{target}}/{{current_version}}",
  TAURI_SIGNING_PRIVATE_KEY: "private-test-key",
  APPLE_SIGNING_IDENTITY: "Developer ID Application: Example",
  APPLE_ID: "release@example.com",
  APPLE_PASSWORD: "test-password",
  APPLE_TEAM_ID: "TEAMID",
};

test("release packaging fails closed without signing, updater or notarization configuration", () => {
  for (const name of [
    "FORTYONE_UPDATER_PUBLIC_KEY",
    "FORTYONE_UPDATER_URL",
    "TAURI_SIGNING_PRIVATE_KEY",
    "APPLE_SIGNING_IDENTITY",
    "APPLE_PASSWORD",
  ]) {
    const env = { ...releaseEnvironment };
    delete env[name];
    assert.throws(() => getReleaseConfiguration(env, "darwin"), /requires/);
  }
  assert.throws(
    () =>
      getReleaseConfiguration(
        { ...releaseEnvironment, APPLE_SIGNING_IDENTITY: "-" },
        "darwin",
      ),
    /ad-hoc/,
  );
  assert.throws(
    () =>
      getReleaseConfiguration(
        {
          ...releaseEnvironment,
          FORTYONE_UPDATER_URL: "http://updates.example.com",
        },
        "darwin",
      ),
    /HTTPS/,
  );
  assert.throws(
    () =>
      getReleaseConfiguration(
        {
          ...releaseEnvironment,
          FORTYONE_UPDATER_URL: "https://secret@example.com",
        },
        "darwin",
      ),
    /credentials/,
  );
});

test("a configured release creates signed updater artifacts without placing secrets in CLI arguments", () => {
  const config = getReleaseConfiguration(releaseEnvironment, "darwin");
  assert.equal(config.bundle.createUpdaterArtifacts, true);
  assert.equal(
    config.bundle.macOS.signingIdentity,
    releaseEnvironment.APPLE_SIGNING_IDENTITY,
  );
  const args = getProductionPackageArguments("darwin", config);
  assert.equal(args.filter((arg) => arg === "--config").length, 2);
  assert.ok(!JSON.stringify(args).includes("private-test-key"));
  assert.ok(!JSON.stringify(args).includes("test-password"));
  const env = getReleaseEnvironment({
    ...releaseEnvironment,
    FORTYONE_DESKTOP_URL: "http://localhost:3000",
    TAURI_CONFIG: "untrusted override",
  });
  assert.equal(
    env.TAURI_SIGNING_PRIVATE_KEY,
    releaseEnvironment.TAURI_SIGNING_PRIVATE_KEY,
  );
  assert.equal(env.FORTYONE_DESKTOP_URL, undefined);
  assert.equal(env.TAURI_CONFIG, undefined);
  assert.equal(
    getProductionEnvironment(releaseEnvironment).TAURI_SIGNING_PRIVATE_KEY,
    undefined,
  );
});

test("Windows release requires a valid installed signing certificate and timestamp service", () => {
  assert.throws(
    () => getReleaseConfiguration(releaseEnvironment, "win32"),
    /CERTIFICATE_THUMBPRINT/,
  );
  const config = getReleaseConfiguration(
    {
      ...releaseEnvironment,
      FORTYONE_WINDOWS_CERTIFICATE_THUMBPRINT: "a".repeat(40),
      FORTYONE_WINDOWS_TIMESTAMP_URL: "http://timestamp.example.com",
    },
    "win32",
  );
  assert.equal(config.bundle.windows.digestAlgorithm, "sha256");
});

test("collaboration CSP only allows the configured websocket origin", () => {
  const config = getFrontendSecurity({
    FORTYONE_COLLABORATION_URL: "wss://collaboration.example.com/socket",
  });
  assert.match(
    config.app.security.csp,
    /connect-src ipc: http:\/\/ipc.localhost wss:\/\/collaboration.example.com;/,
  );
  assert.ok(!config.app.security.csp.includes("wss:" + "*"));
  assert.match(config.app.security.csp, /media-src 'self' blob:;/);
  assert.throws(
    () =>
      getFrontendSecurity({
        FORTYONE_COLLABORATION_URL: "ws://collaboration.example.com",
      }),
    /WSS/,
  );
  assert.throws(
    () =>
      getFrontendSecurity({
        FORTYONE_COLLABORATION_URL:
          "wss://collaboration.example.com?token=secret",
      }),
    /query/,
  );
  assert.match(
    getFrontendSecurity(
      { FORTYONE_COLLABORATION_URL: "ws://localhost:1234" },
      false,
    ).app.security.devCsp,
    /ws:\/\/localhost:1234/,
  );
});
