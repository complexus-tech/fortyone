import assert from "node:assert/strict";
import { test } from "node:test";
import { scrubErrorEvent } from "./observability-privacy.ts";

test("crash diagnostics omit credentials, document content, and personal context", () => {
  const event = scrubErrorEvent({
    type: undefined,
    event_id: "a".repeat(32),
    release: "com.fortyone.mobile@1.0.0+1",
    message: "Authorization: Bearer private-token",
    request: {
      url: "https://api.example.invalid/stories?token=private-token",
      headers: { Cookie: "fortyone_session=private-cookie" },
      data: "<p>Private document</p>",
    },
    user: { email: "private@example.invalid", id: "private-account" },
    extra: { body: "Private document", password: "private-password" },
    breadcrumbs: [{ message: "Private document" }],
    contexts: {
      device: { name: "Private person's phone", model: "iPhone" },
      document: { html: "<p>Private document</p>" },
    },
    exception: {
      values: [
        {
          type: "TypeError",
          value: "Server returned <p>Private document</p>",
          mechanism: {
            type: "generic",
            handled: true,
            data: { body: "Private document" },
          },
          stacktrace: {
            frames: [
              {
                filename:
                  "https://user:password@cdn.example.invalid/main.js?token=private-token#secret",
                function: "saveStory",
                lineno: 24,
                colno: 8,
                context_line: "Private document",
                vars: { cookie: "private-cookie" },
              },
            ],
          },
        },
      ],
    },
  });

  assert.doesNotMatch(JSON.stringify(event), /private|password|secret|<p>/i);
  assert.equal(event.exception?.values?.[0].type, "TypeError");
  assert.deepEqual(event.exception?.values?.[0].stacktrace?.frames?.[0], {
    filename: "https://cdn.example.invalid/main.js",
    abs_path: undefined,
    function: "saveStory",
    module: undefined,
    lineno: 24,
    colno: 8,
    in_app: undefined,
    instruction_addr: undefined,
    platform: undefined,
  });
  assert.equal(event.contexts?.device?.model, "iPhone");
});

test("native thread stacks retain addresses while removing local usernames and stack variables", () => {
  const event = scrubErrorEvent({
    type: undefined,
    threads: {
      values: [
        {
          id: 1,
          crashed: true,
          name: "Private thread",
          stacktrace: {
            frames: [
              {
                filename: "/Users/private-person/build/main.js",
                instruction_addr: "0x1234",
                vars: { html: "Private document" },
              },
            ],
          },
        },
      ],
    },
  });
  assert.doesNotMatch(JSON.stringify(event), /private/i);
  assert.equal(
    event.threads?.values?.[0].stacktrace?.frames?.[0].filename,
    "main.js",
  );
  assert.equal(
    event.threads?.values?.[0].stacktrace?.frames?.[0].instruction_addr,
    "0x1234",
  );
});
