import { waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { getDesktopBridge } from "./bridge";
import { navigateWithDesktopOffline } from "./offline-navigation";

jest.mock("sonner", () => ({
  toast: {
    loading: jest.fn(() => "progress"),
    error: jest.fn(),
    dismiss: jest.fn(),
  },
}));
jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));

const originalLocation = window.location;
const assign = jest.fn();
const restoreSession = jest.fn();
const pageRead = jest.fn();
const scope = {
  accountId: "account",
  workspaceId: "workspace",
  workspaceSlug: "engineering",
  expiresAt: Date.now() + 100_000,
};
let online = false;
let listener: EventListener | undefined;
beforeEach(() => {
  jest.clearAllMocks();
  online = false;
  delete window.__FORTYONE_OFFLINE_PAGE__;
  jest.spyOn(navigator, "onLine", "get").mockImplementation(() => online);
  Object.defineProperty(window, "location", {
    configurable: true,
    value: {
      href: "http://localhost/engineering/my-work",
      origin: "http://localhost",
      hostname: "localhost",
      pathname: "/engineering/my-work",
      assign,
    },
  });
  assign.mockImplementation(() => {
    window.dispatchEvent(new Event("pagehide"));
  });
  restoreSession.mockResolvedValue(scope);
  pageRead.mockResolvedValue({ html: "saved", savedAt: Date.now() });
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ restoreSession, pageRead } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
  document.body.inert = false;
});
afterEach(() => {
  if (listener)
    window.removeEventListener("fortyone:desktop-before-close", listener);
  listener = undefined;
  Object.defineProperty(window, "location", {
    configurable: true,
    value: originalLocation,
  });
  jest.restoreAllMocks();
  jest.useRealTimers();
});
const beforeNavigate = (write: Promise<void>) => {
  listener = (event) => {
    (
      event as CustomEvent<{ waitUntil: (write: Promise<void>) => void }>
    ).detail.waitUntil(write);
  };
  window.addEventListener("fortyone:desktop-before-close", listener);
};

it("keeps ordinary online navigation on the existing router", async () => {
  online = true;
  const fallback = jest.fn();
  await navigateWithDesktopOffline("/engineering/docs", fallback);
  expect(fallback).toHaveBeenCalledTimes(1);
  expect(pageRead).not.toHaveBeenCalled();
  expect(assign).not.toHaveBeenCalled();
});

it("coalesces clicks and waits for editor writes before navigating", async () => {
  let saved!: () => void;
  beforeNavigate(
    new Promise((resolve) => {
      saved = resolve;
    }),
  );
  const first = navigateWithDesktopOffline("/engineering/docs", jest.fn());
  const second = navigateWithDesktopOffline("/engineering/settings", jest.fn());
  expect(first).toBe(second);
  await waitFor(() => {
    expect(pageRead).toHaveBeenCalledTimes(1);
  });
  expect(document.body.inert).toBe(true);
  expect(assign).not.toHaveBeenCalled();
  expect(toast.loading).not.toHaveBeenCalled();
  saved();
  await first;
  expect(assign).toHaveBeenCalledWith("http://localhost/engineering/docs");
  expect(document.body.inert).toBe(false);
  expect(toast.error).not.toHaveBeenCalled();
});

it("keeps the current editor open when a local write fails", async () => {
  beforeNavigate(Promise.reject(new Error("Disk full")));
  await navigateWithDesktopOffline("/engineering/docs", jest.fn());
  expect(assign).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenCalledWith(
    "Could not open the downloaded page",
    { description: "Disk full" },
  );
  expect(document.body.inert).toBe(false);
});

it("does not leave the current page for an undownloaded route or changed account", async () => {
  pageRead.mockResolvedValueOnce(null);
  await navigateWithDesktopOffline("/engineering/docs", jest.fn());
  expect(assign).not.toHaveBeenCalled();
  restoreSession
    .mockResolvedValueOnce(scope)
    .mockResolvedValueOnce({ ...scope, accountId: "different" });
  await navigateWithDesktopOffline("/engineering/docs", jest.fn());
  expect(assign).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenLastCalledWith(
    "Could not open the downloaded page",
    expect.objectContaining({
      description: expect.stringContaining("account or workspace changed"),
    }),
  );
});

it("unfreezes a canceled navigation after a bounded wait", async () => {
  jest.useFakeTimers();
  const stop = jest.spyOn(window, "stop").mockImplementation(() => undefined);
  assign.mockImplementation(() => undefined);
  const task = navigateWithDesktopOffline("/engineering/docs", jest.fn());
  await jest.advanceTimersByTimeAsync(0);
  expect(assign).toHaveBeenCalledTimes(1);
  expect(document.body.inert).toBe(true);
  await jest.advanceTimersByTimeAsync(10_000);
  await task;
  expect(stop).toHaveBeenCalledTimes(1);
  expect(document.body.inert).toBe(false);
});

it("uses the downloaded page and draft gate when the network is online but the server is unavailable", async () => {
  online = true;
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  const fallback = jest.fn();
  const flush = jest.fn();
  listener = (event) => {
    flush();
    (
      event as CustomEvent<{ waitUntil: (write: Promise<void>) => void }>
    ).detail.waitUntil(Promise.resolve());
  };
  window.addEventListener("fortyone:desktop-before-close", listener);
  await navigateWithDesktopOffline("/engineering/settings", fallback);
  expect(flush).toHaveBeenCalledTimes(1);
  expect(fallback).not.toHaveBeenCalled();
  expect(assign).toHaveBeenCalledWith("http://localhost/engineering/settings");
});
