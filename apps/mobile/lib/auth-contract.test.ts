import assert from "node:assert/strict";
import { test } from "node:test";
import { parseSessionCookie, readMobileCallbackCode } from "./auth-contract.ts";

const value = "x".repeat(43);
const cookie = `fortyone_session=${value}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=3600; Domain=.fortyone.app`;
const apiURL = new URL("https://api.fortyone.app");
const affinityCookie =
  "AWSALBTG=load-balancer-affinity; Expires=Tue, 22 Sep 2026 12:00:00 GMT; Path=/";
const corsAffinityCookie =
  "AWSALBTGCORS=load-balancer-affinity; Expires=Tue, 22 Sep 2026 12:00:00 GMT; Path=/; SameSite=None; Secure";

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

test("selects the session from combined iOS headers without splitting expiry dates", () => {
  const sessionCookie = cookie.replace(
    "Path=/;",
    "Path=/; Expires=Thu, 15 Oct 2026 12:00:00 GMT;",
  );
  for (const cookies of [
    [affinityCookie, corsAffinityCookie, sessionCookie],
    [affinityCookie, sessionCookie, corsAffinityCookie],
    [sessionCookie, affinityCookie, corsAffinityCookie],
  ]) {
    for (const separator of [", ", ","]) {
      assert.deepEqual(
        parseSessionCookie(cookies.join(separator), apiURL, 1000),
        { cookie: `fortyone_session=${value}`, expiresAt: 3_601_000 },
      );
    }
  }
});

test("accepts separate Set-Cookie values and combined values in the same response", () => {
  assert.deepEqual(
    parseSessionCookie(
      [affinityCookie, `${cookie}, ${corsAffinityCookie}`],
      apiURL,
      1000,
    ),
    { cookie: `fortyone_session=${value}`, expiresAt: 3_601_000 },
  );
});

test("rejects missing or ambiguous session cookies among unrelated cookies", () => {
  for (const header of [
    `${affinityCookie}, ${corsAffinityCookie}`,
    cookie.replace("fortyone_session=", "fortyone_session_other="),
    `${cookie}, ${cookie}`,
    [cookie, cookie.replace(value, "y".repeat(43))],
    [],
  ]) {
    assert.throws(() => parseSessionCookie(header, apiURL));
  }
});

test("validates only the selected session's token and cookie policy", () => {
  for (const invalid of [
    cookie.replace(value, "short"),
    cookie.replace("Secure; ", ""),
    cookie.replace("HttpOnly; ", ""),
    cookie.replace("Max-Age=3600", "Max-Age=0"),
    cookie.replace("Domain=.fortyone.app", "Domain=attacker.example"),
  ]) {
    assert.throws(() =>
      parseSessionCookie(
        `${affinityCookie}, ${invalid}, ${cookie.replace("fortyone_session=", "other=")}`,
        apiURL,
      ),
    );
  }
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
