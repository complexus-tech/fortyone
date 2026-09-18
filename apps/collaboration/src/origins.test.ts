import { strict as assert } from "node:assert";
import { test } from "node:test";
import { isAllowedOrigin } from "./origins";

test("workspace origins match a bounded HTTPS suffix", () => {
  const allowed = ["https://*.fortyone.app", "http://localhost:3000"];
  assert.equal(isAllowedOrigin("https://acme.fortyone.app", allowed), true);
  assert.equal(isAllowedOrigin("http://localhost:3000", allowed), true);
  for (const origin of [
    "",
    "null",
    "https://fortyone.app.evil.com",
    "https://evilfortyone.app",
    "http://acme.fortyone.app",
    "https://acme.fortyone.app:444",
    "https://acme.fortyone.app/path",
  ])
    assert.equal(isAllowedOrigin(origin, allowed), false, origin);
});
