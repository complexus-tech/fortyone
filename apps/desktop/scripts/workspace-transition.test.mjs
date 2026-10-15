import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { runInNewContext } from "node:vm";

const source = readFileSync(
  new URL("../src-tauri/src/workspace-transition.js", import.meta.url),
  "utf8",
);
const ROOT_ID = "fortyone-native-workspace-switch";

// The injected script deliberately uses only this small, standard DOM surface.
// No renderer, network, Tauri runtime, or workspace dependencies are needed.
class Element extends EventTarget {
  constructor(tag) {
    super();
    this.tagName = tag;
    this.children = [];
    this.dataset = {};
    this.attributes = {};
    this.textContent = "";
  }
  setAttribute(name, value) {
    this.attributes[name] = value;
  }
  appendChild(child) {
    child.parent = this;
    this.children.push(child);
    return child;
  }
  append(...children) {
    for (const child of children) this.appendChild(child);
  }
  remove() {
    if (this.parent)
      this.parent.children = this.parent.children.filter(
        (child) => child !== this,
      );
  }
  querySelectorAll(selector) {
    const matches = (element) =>
      selector === "button"
        ? element.tagName === "button"
        : selector.startsWith(".")
          ? element.className === selector.slice(1)
          : selector.startsWith("#")
            ? element.id === selector.slice(1)
            : false;
    return this.children.flatMap((child) => [
      ...(matches(child) ? [child] : []),
      ...child.querySelectorAll(selector),
    ]);
  }
  querySelector(selector) {
    return this.querySelectorAll(selector)[0] ?? null;
  }
  get innerHTML() {
    throw new Error("The progress overlay must never parse user-supplied HTML");
  }
  set innerHTML(_) {
    throw new Error("The progress overlay must never parse user-supplied HTML");
  }
}
const flush = async () => {
  for (let turn = 0; turn < 8; turn++) await Promise.resolve();
};
function page({
  frame = false,
  loading = false,
  state = null,
  sourceProgress = false,
} = {}) {
  const body = new Element("body");
  const head = new Element("head");
  const timers = new Map();
  const calls = [];
  const errors = [];
  let sequence = 0;
  let handler = async (command) => {
    if (command === "cancel_workspace_switch") state = null;
    if (command === "retry_workspace_switch")
      state = { ...state, failed: false, startedAt: 2 };
    return command === "get_workspace_switch" ? state : undefined;
  };
  const document = Object.assign(new EventTarget(), {
    body: loading ? null : body,
    head,
    readyState: loading ? "loading" : "complete",
    createElement: (tag) => new Element(tag),
    getElementById: (id) =>
      body.querySelector(`#${id}`) ?? head.querySelector(`#${id}`),
    querySelector: () => (sourceProgress ? new Element("button") : null),
  });
  const window = Object.assign(new EventTarget(), {
    __TAURI_INTERNALS__: {
      invoke: (command) => {
        calls.push(command);
        return handler(command);
      },
    },
  });
  window.top = frame ? {} : window;
  runInNewContext(source, {
    window,
    document,
    setTimeout: (callback, milliseconds) => {
      const id = ++sequence;
      timers.set(id, { callback, milliseconds });
      return id;
    },
    clearTimeout: (id) => timers.delete(id),
    console: { error: (...args) => errors.push(args) },
  });
  return {
    window,
    document,
    body,
    timers,
    calls,
    errors,
    root: () => document.getElementById(ROOT_ID),
    setState: (next) => {
      state = next;
    },
    setHandler: (next) => {
      handler = next;
    },
    setSourceProgress: (next) => {
      sourceProgress = next;
    },
    refresh: () =>
      window.dispatchEvent(new Event("fortyone:workspace-switch-change")),
  };
}
const destination = {
  id: "workspace",
  name: "Product",
  startedAt: 1,
  failed: false,
};

test("restores native state on the new document and treats workspace names as text", async () => {
  const view = page({
    state: { ...destination, name: "<img src=x onerror=alert(1)>" },
  });
  await flush();
  assert.equal(view.root().attributes.role, "status");
  assert.equal(
    view.root().querySelector(".workspace-switch-heading").children[1]
      .textContent,
    "Opening <img src=x onerror=alert(1)>…",
  );
  assert.equal(view.timers.size, 1);
  assert.equal([...view.timers.values()][0].milliseconds, 500);
});

test("leaves idle documents alone and clears completed transitions", async () => {
  const view = page();
  await flush();
  assert.equal(view.root(), null);
  assert.equal(view.timers.size, 0);
  view.setState(destination);
  view.refresh();
  await flush();
  assert.ok(view.root());
  view.setState(null);
  view.refresh();
  await flush();
  assert.equal(view.root(), null);
  assert.equal(view.timers.size, 0);
});

test("suppresses duplicate feedback while the source React switcher is present", async () => {
  const view = page({ state: destination, sourceProgress: true });
  await flush();
  assert.equal(view.root(), null);
  view.setSourceProgress(false);
  view.refresh();
  await flush();
  assert.ok(view.root());
  const first = view.root();
  view.refresh();
  await flush();
  assert.equal(
    view.root(),
    first,
    "polling must preserve focus and DOM identity",
  );
});

test("timeout actions call native recovery without navigating to supplied URLs", async () => {
  const view = page({
    state: { ...destination, failed: true, url: "https://attacker.invalid" },
  });
  await flush();
  assert.equal(view.root().attributes.role, "alert");
  view.root().querySelectorAll("button")[0].dispatchEvent(new Event("click"));
  await flush();
  assert.ok(view.calls.includes("retry_workspace_switch"));
  assert.equal(view.root().attributes.role, "status");
  view.setState({ ...destination, failed: true });
  view.refresh();
  await flush();
  view.root().querySelectorAll("button")[1].dispatchEvent(new Event("click"));
  await flush();
  assert.ok(view.calls.includes("cancel_workspace_switch"));
  assert.equal(view.root(), null);
});

test("recovery failure remains actionable and reports connection trouble", async () => {
  const view = page({ state: { ...destination, failed: true } });
  await flush();
  view.setHandler(async () => {
    throw new Error("offline");
  });
  view.root().querySelectorAll("button")[0].dispatchEvent(new Event("click"));
  await flush();
  assert.ok(
    view
      .root()
      .querySelectorAll("button")
      .every((button) => !button.disabled),
  );
  assert.match(
    view.root().querySelector(".workspace-switch-description").textContent,
    /Check your connection/,
  );
  assert.equal(view.errors.length, 1);
});

test("pagehide cancels polling and ignores an in-flight stale response", async () => {
  const view = page();
  await flush();
  let resolve;
  view.setHandler(
    () =>
      new Promise((done) => {
        resolve = done;
      }),
  );
  view.refresh();
  view.window.dispatchEvent(new Event("pagehide"));
  resolve(destination);
  await flush();
  assert.equal(view.root(), null);
  assert.equal(view.timers.size, 0);
});

test("waits for document parsing and never runs in child frames", async () => {
  const view = page({ state: destination, loading: true });
  await flush();
  assert.equal(view.root(), null);
  view.document.body = view.body;
  view.document.dispatchEvent(new Event("DOMContentLoaded"));
  await flush();
  assert.ok(view.root());
  const child = page({ state: destination, frame: true });
  await flush();
  assert.deepEqual(child.calls, []);
  assert.equal(child.root(), null);
});
