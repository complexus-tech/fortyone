import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import test from "node:test";
import { fileURLToPath } from "node:url";
import {
  getDesktopUrl,
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
  assert.equal(args[args.indexOf("--features") + 1], "production");
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
  assert.deepEqual(config.build, {
    devUrl: null,
    frontendDist: "https://cloud.fortyone.app",
  });
  assert.equal(config.bundle.active, true);
  assert.deepEqual(config.bundle.targets, ["dmg"]);
  assert.equal(config.bundle.macOS.signingIdentity, "-");

  assert.throws(() => getProductionPackageArguments("linux"), /on a Mac/);
  assert.throws(() => getProductionPackageArguments("win32"), /on a Mac/);
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
