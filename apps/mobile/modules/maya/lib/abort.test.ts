import assert from "node:assert/strict";
import { createRequire } from "node:module";
import test from "node:test";
import { assertMayaRequestNotAborted } from "./abort.ts";

// Use the exact dependency installed by React Native's setUpXHR.js.
const require = createRequire(import.meta.url);
const requireFromReactNative = createRequire(
  require.resolve("react-native/package.json"),
);
const { AbortController: NativeAbortController } = requireFromReactNative(
  "abort-controller/dist/abort-controller",
) as {
  AbortController: new () => {
    signal: { readonly aborted: boolean; readonly reason?: unknown };
    abort: () => void;
  };
};

test("the installed React Native signal works without throwIfAborted", () => {
  const controller = new NativeAbortController();
  assert.equal("throwIfAborted" in controller.signal, false);
  assert.doesNotThrow(() => assertMayaRequestNotAborted(controller.signal));
});

test("an already cancelled native signal throws a useful AbortError", () => {
  const controller = new NativeAbortController();
  controller.abort();
  assert.throws(() => assertMayaRequestNotAborted(controller.signal), {
    name: "AbortError",
    message: "Maya request cancelled.",
  });
});

test("native cancellation during a context read prevents subsequent dispatch", async () => {
  const controller = new NativeAbortController();
  let finishContext!: () => void;
  const context = new Promise<void>((resolve) => {
    finishContext = resolve;
  });
  let sent = false;
  const prepareAndSend = async () => {
    assertMayaRequestNotAborted(controller.signal);
    await context;
    assertMayaRequestNotAborted(controller.signal);
    sent = true;
  };
  const request = prepareAndSend();
  controller.abort();
  finishContext();
  await assert.rejects(request, { name: "AbortError" });
  assert.equal(sent, false);
});

test("a supported cancellation reason is preserved exactly", () => {
  const reason = new Error("Session changed");
  assert.throws(
    () => assertMayaRequestNotAborted({ aborted: true, reason }),
    (error: unknown) => error === reason,
  );
});
