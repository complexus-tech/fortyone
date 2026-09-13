import assert from "node:assert/strict";
import { test } from "node:test";
import { assertApiRequestURL, getApiURL } from "./config.ts";

test("credentials are restricted to the configured API origin and base path", () => {
  const previous = process.env.EXPO_PUBLIC_API_URL;
  process.env.EXPO_PUBLIC_API_URL = "https://api.fortyone.app/v1";
  try {
    assert.doesNotThrow(() =>
      assertApiRequestURL("https://api.fortyone.app/v1/workspaces"),
    );
    for (const target of [
      "https://attacker.example/v1/workspaces",
      "https://api.fortyone.app/v10/workspaces",
      "https://api.fortyone.app/other",
      "https://user:password@api.fortyone.app/v1/workspaces",
    ]) {
      assert.throws(
        () => assertApiRequestURL(target),
        /cannot send credentials/,
      );
    }
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_API_URL;
    else process.env.EXPO_PUBLIC_API_URL = previous;
  }
});

test("rejects insecure and credential-bearing API configuration", () => {
  const previous = process.env.EXPO_PUBLIC_API_URL;
  try {
    for (const value of [
      "http://api.fortyone.app",
      "https://user:password@api.fortyone.app",
      "https://api.fortyone.app?token=unsafe",
    ]) {
      process.env.EXPO_PUBLIC_API_URL = value;
      assert.throws(() => getApiURL(), /trusted HTTPS URL/);
    }
  } finally {
    if (previous === undefined) delete process.env.EXPO_PUBLIC_API_URL;
    else process.env.EXPO_PUBLIC_API_URL = previous;
  }
});
