import type * as ActiveViewModule from "./active-view";

type ActiveView = typeof ActiveViewModule;

const newTab = () => {
  let tab: ActiveView | undefined;
  jest.isolateModules(() => {
    tab = jest.requireActual<ActiveView>("./active-view");
  });
  if (!tab) throw new Error("Expected a tab visibility tracker");
  return tab;
};

beforeEach(() => {
  jest.useFakeTimers();
  localStorage.clear();
});
afterEach(() => {
  jest.clearAllTimers();
  jest.useRealTimers();
});

it("preserves recently viewed messages from multiple tabs without overwriting them", () => {
  const first = newTab();
  const second = newTab();
  first.recordChatMessageVisible("alice", "Engineering", { id: "first" });
  second.recordChatMessageVisible("alice", "engineering", { id: "second" });
  jest.advanceTimersByTime(100);
  expect(first.wasChatMessageVisible("alice", "engineering", "second")).toBe(
    true,
  );
  expect(second.wasChatMessageVisible("alice", "engineering", "first")).toBe(
    true,
  );
  expect(second.wasChatMessageVisible("bob", "engineering", "first")).toBe(
    false,
  );
  expect(second.wasChatMessageVisible("alice", "marketing", "first")).toBe(
    false,
  );
});

it("expires old view evidence instead of suppressing future notifications indefinitely", () => {
  const first = newTab();
  const second = newTab();
  first.recordChatMessageVisible("alice", "engineering", { id: "first" });
  jest.advanceTimersByTime(100);
  expect(second.wasChatMessageVisible("alice", "engineering", "first")).toBe(
    true,
  );
  jest.advanceTimersByTime(10_000);
  expect(second.wasChatMessageVisible("alice", "engineering", "first")).toBe(
    false,
  );
});
