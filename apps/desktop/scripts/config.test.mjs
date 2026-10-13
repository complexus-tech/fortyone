import assert from "node:assert/strict";
import test from "node:test";
import { getDesktopUrl, getTauriArguments } from "./config.mjs";

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
