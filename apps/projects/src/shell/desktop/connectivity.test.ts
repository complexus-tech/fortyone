import { act } from "@testing-library/react";
import { getDesktopBridge } from "./bridge";
import {
  DESKTOP_CONNECTIVITY_EVENT,
  getDesktopOnline,
  monitorDesktopConnectivity,
  probeDesktopConnectivity,
} from "./connectivity";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "http://localhost:8000",
}));
const request = jest.fn();
let stop: (() => void) | undefined;
beforeEach(() => {
  jest.useFakeTimers();
  jest.clearAllMocks();
  delete window.__FORTYONE_OFFLINE_PAGE__;
  jest.spyOn(navigator, "onLine", "get").mockReturnValue(true);
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({} as NonNullable<ReturnType<typeof getDesktopBridge>>);
  Object.assign(globalThis, { fetch: request });
  request.mockResolvedValue({ ok: true });
});
afterEach(() => {
  stop?.();
  stop = undefined;
  delete window.__FORTYONE_OFFLINE_PAGE__;
  jest.restoreAllMocks();
  jest.useRealTimers();
});

it("starts cached pages offline even when the device has a network connection", async () => {
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  request.mockRejectedValue(new TypeError("Failed to fetch"));
  expect(getDesktopOnline()).toBe(false);
  stop = monitorDesktopConnectivity();
  await act(async () => {
    await jest.advanceTimersByTimeAsync(0);
  });
  expect(request).toHaveBeenCalledTimes(2);
  expect((request.mock.calls[0][0] as URL).pathname).toBe("/desktop-worker.js");
  expect(getDesktopOnline()).toBe(false);
  expect(request).toHaveBeenNthCalledWith(
    1,
    expect.any(URL),
    expect.objectContaining({ method: "HEAD", cache: "no-store" }),
  );
  expect(request).toHaveBeenNthCalledWith(
    2,
    expect.any(URL),
    expect.objectContaining({ method: "GET", cache: "no-store" }),
  );
});

it("recovers only when both services respond, then stops periodic probing", async () => {
  window.__FORTYONE_OFFLINE_PAGE__ = true;
  request
    .mockResolvedValueOnce({ ok: true })
    .mockResolvedValueOnce({ ok: false });
  const changed = jest.fn();
  window.addEventListener(DESKTOP_CONNECTIVITY_EVENT, changed);
  stop = monitorDesktopConnectivity();
  await act(async () => {
    await jest.advanceTimersByTimeAsync(0);
  });
  expect(getDesktopOnline()).toBe(false);
  await act(async () => {
    await jest.advanceTimersByTimeAsync(20_000);
  });
  expect(getDesktopOnline()).toBe(true);
  expect(changed).toHaveBeenCalledTimes(1);
  const count = request.mock.calls.length;
  await act(async () => {
    await jest.advanceTimersByTimeAsync(60_000);
  });
  expect(request).toHaveBeenCalledTimes(count);
  window.removeEventListener(DESKTOP_CONNECTIVITY_EVENT, changed);
});

it("coalesces probes and never probes ordinary web browsers", async () => {
  let finish!: (response: { ok: boolean }) => void;
  request.mockImplementation(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  // Return a shared response promise so both bounded probes finish together.
  const response = new Promise((resolve) => {
    finish = resolve;
  });
  request.mockReturnValue(response);
  const first = probeDesktopConnectivity();
  expect(probeDesktopConnectivity()).toBe(first);
  finish({ ok: true });
  await first;
  request.mockClear();
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  stop = monitorDesktopConnectivity();
  await probeDesktopConnectivity();
  expect(request).not.toHaveBeenCalled();
});
