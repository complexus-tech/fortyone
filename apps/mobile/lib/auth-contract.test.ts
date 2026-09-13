import assert from "node:assert/strict";
import { test } from "node:test";
import { parseSessionCookie, readMobileCallbackCode } from "./auth-contract.ts";

const value = "x".repeat(43);
const cookie = `fortyone_session=${value}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=3600; Domain=.fortyone.app`;
const apiURL = new URL("https://api.fortyone.app");

test("keeps only the expected cookie and authoritative relative expiry", () => {
  assert.deepEqual(parseSessionCookie(cookie, apiURL, 1000), {
    cookie: `fortyone_session=${value}`,
    expiresAt: 3_601_000,
  });
});

test("rejects malformed, expired, cross-domain and insecure session cookies", () => {
  for (const header of [
    null,
    "",
    cookie.replace(value, "short"),
    cookie.replace("Max-Age=3600", "Max-Age=-1"),
    cookie.replace("HttpOnly; ", ""),
    cookie.replace("Secure; ", ""),
    cookie.replace("Domain=.fortyone.app", "Domain=attacker.example"),
    cookie.replace("Path=/;", "Path=/different;"),
  ])
    assert.throws(() => parseSessionCookie(header, apiURL));
});

test("accepts only the fixed callback with exactly the expected state", () => {
  const state = "s".repeat(43);
  const callback = `fortyone://login?code=${value}&state=${state}`;
  assert.equal(readMobileCallbackCode(callback, state), value);
  for (const invalid of [
    callback.replace("fortyone:", "attacker:"),
    callback.replace("login?", "login/extra?"),
    `${callback}&state=${state}`,
    `${callback}&code=${value}`,
    `${callback}#fragment`,
    callback.replace(state, "z".repeat(43)),
  ])
    assert.throws(() => readMobileCallbackCode(invalid, state));
});
