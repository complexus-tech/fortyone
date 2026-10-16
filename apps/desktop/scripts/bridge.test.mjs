import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { runInNewContext } from "node:vm";

const bridgeSource = readFileSync(
  new URL("../src-tauri/src/bridge.js", import.meta.url),
  "utf8",
);

function createPage({ readyState = "loading", frame = false, platform = "macos" } = {}) {
  const styles = [];
  const historyCalls = [];
  const nativeCalls = [];
  const document = Object.assign(new EventTarget(), {
    readyState,
    body: { inert: false },
    createElement(tag) {
      assert.equal(tag, "style");
      return { textContent: "" };
    },
    head: {
      appendChild(style) {
        styles.push(style);
        return style;
      },
    },
  });
  const window = Object.assign(new EventTarget(), {
    history: Object.fromEntries(
      ["pushState", "replaceState"].map((method) => [
        method,
        function (...args) {
          assert.equal(this, window.history);
          historyCalls.push({ method, args });
          return method;
        },
      ]),
    ),
    __TAURI_INTERNALS__: {
      invoke(command, args) {
        nativeCalls.push({ command, args });
        return Promise.resolve(command);
      },
    },
  });
  window.top = frame ? {} : window;

  return {
    window,
    document,
    styles,
    historyCalls,
    nativeCalls,
    initialize: () =>
      runInNewContext(
        bridgeSource.replaceAll("__FORTYONE_DESKTOP_PLATFORM__", platform),
        { window, document, Event, CustomEvent },
      ),
  };
}

test("the document-start bridge waits for the head and adds its stylesheet once", () => {
  const page = createPage();
  page.initialize();
  assert.equal(page.styles.length, 0);
  assert.ok(page.window.__FORTYONE_DESKTOP__);

  page.document.dispatchEvent(new Event("DOMContentLoaded"));
  assert.equal(page.styles.length, 1);
  assert.ok(page.styles[0].textContent.trim());

  page.document.dispatchEvent(new Event("DOMContentLoaded"));
  assert.equal(page.styles.length, 1);
});

test("an already parsed document receives the stylesheet immediately", () => {
  for (const readyState of ["interactive", "complete"]) {
    const page = createPage({ readyState });
    page.initialize();
    assert.equal(page.styles.length, 1, readyState);
    page.document.dispatchEvent(new Event("DOMContentLoaded"));
    assert.equal(page.styles.length, 1, readyState);
  }
});

test("child frames receive neither chrome styles nor desktop bridge access", () => {
  const page = createPage({ frame: true, readyState: "complete" });
  const { pushState, replaceState } = page.window.history;
  page.initialize();
  page.document.dispatchEvent(new Event("DOMContentLoaded"));

  assert.equal(page.styles.length, 0);
  assert.equal(page.window.__FORTYONE_DESKTOP__, undefined);
  assert.equal(page.window.history.pushState, pushState);
  assert.equal(page.window.history.replaceState, replaceState);
});

test("history notifications preserve Next metadata, arguments, and return values", () => {
  const page = createPage();
  const metadata = Object.freeze({
    __NA: true,
    __PRIVATE_NEXTJS_INTERNALS_TREE: Object.freeze(["workspace", "stories"]),
  });
  const notificationOrder = [];
  page.window.addEventListener("fortyone:desktop-history-change", () => {
    notificationOrder.push(page.historyCalls.length);
  });
  page.initialize();

  for (const method of ["pushState", "replaceState"]) {
    assert.equal(
      page.window.history[method](metadata, "", "/workspace/stories"),
      method,
    );
  }

  assert.deepEqual(notificationOrder, [1, 2]);
  assert.deepEqual(
    page.historyCalls.map(({ method }) => method),
    ["pushState", "replaceState"],
  );
  for (const { args } of page.historyCalls) {
    assert.equal(args[0], metadata);
    assert.deepEqual(args.slice(1), ["", "/workspace/stories"]);
  }
});

test("a rejected history update propagates without notifying navigation", () => {
  const page = createPage();
  const failure = new Error("History update denied");
  page.window.history.pushState = () => {
    throw failure;
  };
  let notifications = 0;
  page.window.addEventListener("fortyone:desktop-history-change", () => {
    notifications += 1;
  });
  page.initialize();

  assert.throws(() => page.window.history.pushState({}, "", "/"), failure);
  assert.equal(notifications, 0);
});

test("the immutable version-two bridge retains its four native command contracts", async () => {
  const page = createPage();
  page.initialize();
  const bridge = page.window.__FORTYONE_DESKTOP__;
  const state = { ready: false, background: [255, 255, 255] };
  assert.equal(bridge.version, 2);
  assert.equal(bridge.platform, "macos");
  assert.equal(bridge.vibrancy, true);
  assert.ok(Object.isFrozen(bridge));
  const descriptor = Object.getOwnPropertyDescriptor(
    page.window,
    "__FORTYONE_DESKTOP__",
  );
  assert.equal(descriptor.writable, false);
  assert.equal(descriptor.configurable, false);

  assert.equal(await bridge.syncChrome(state), "sync_desktop_chrome");
  assert.equal(await bridge.navigationState(), "desktop_navigation_state");
  assert.equal(await bridge.navigate("back"), "desktop_navigate");
  assert.equal(await bridge.windowAction("drag"), "desktop_window_action");
  assert.deepEqual(
    page.nativeCalls.map(({ command }) => command),
    [
      "sync_desktop_chrome",
      "desktop_navigation_state",
      "desktop_navigate",
      "desktop_window_action",
    ],
  );
  assert.equal(page.nativeCalls[0].args.state, state);
  assert.equal(page.nativeCalls[1].args, undefined);
  assert.equal(page.nativeCalls[2].args.direction, "back");
  assert.equal(page.nativeCalls[3].args.action, "drag");
});

test("the optional menu contract syncs presentation separately from window chrome", async () => {
  const page = createPage();
  page.initialize();
  const state = {
    ready: true,
    canCreateStory: true,
    canCreateObjective: false,
    canToggleSidebar: false,
    sidebarCollapsed: true,
    storyLabel: "Ticket",
    objectiveLabel: "Goal",
  };

  assert.equal(
    await page.window.__FORTYONE_DESKTOP__.syncMenu(state),
    "sync_desktop_menu",
  );
  assert.equal(page.nativeCalls.length, 1);
  assert.equal(page.nativeCalls[0].args.state, state);
});

test("Windows and Linux retain native capabilities and await draft saving before hiding", async () => {
  for (const platform of ["windows", "linux"]) {
    const page = createPage({ platform });
    let saved;
    const save = new Promise((resolve) => { saved = resolve; });
    page.window.addEventListener("fortyone:desktop-before-close", (event) => {
      event.detail.waitUntil(save);
    });
    page.initialize();
    assert.equal(page.window.__FORTYONE_DESKTOP__.platform, platform);
    assert.equal(page.window.__FORTYONE_DESKTOP__.vibrancy, false);
    assert.equal(await page.window.__FORTYONE_DESKTOP__.getPreferences(), "hosted_get_preferences");

    page.window.dispatchEvent(new CustomEvent("fortyone:native-before-close", {
      detail: { nonce: "close-request", action: "hide" },
    }));
    assert.equal(page.document.body.inert, true);
    assert.equal(page.nativeCalls.some(({ command }) => command === "hosted_confirm_close"), false);
    saved();
    await new Promise(setImmediate);
    const confirmation = page.nativeCalls.find(({ command }) => command === "hosted_confirm_close");
    assert.equal(confirmation.args.nonce, "close-request");
    assert.equal(confirmation.args.allow, true);
    assert.equal(page.document.body.inert, false);
  }
});
