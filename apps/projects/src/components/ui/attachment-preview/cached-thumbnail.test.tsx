/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals. */
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { CachedAttachmentThumbnail } from "./cached-thumbnail";

const mockLoad = jest.fn();
const mockCache = {
  load: (...args: unknown[]) => mockLoad(...args),
  subscribe: () => () => undefined,
  revision: () => 0,
};
jest.mock("./thumbnail-cache-context", () => ({
  useAttachmentThumbnailCache: () => mockCache,
}));
const originalCreate = Object.getOwnPropertyDescriptor(URL, "createObjectURL");
const originalRevoke = Object.getOwnPropertyDescriptor(URL, "revokeObjectURL");
beforeEach(() => {
  mockLoad.mockReset().mockImplementation(async () => ({
    blob: new Blob(["jpeg"], { type: "image/jpeg" }),
    checkedAt: Date.now(),
  }));
  URL.createObjectURL = jest.fn(() => "blob:preview");
  URL.revokeObjectURL = jest.fn();
});
afterEach(() => {
  cleanup();
  if (originalCreate)
    Object.defineProperty(URL, "createObjectURL", originalCreate);
  else Reflect.deleteProperty(URL, "createObjectURL");
  if (originalRevoke)
    Object.defineProperty(URL, "revokeObjectURL", originalRevoke);
  else Reflect.deleteProperty(URL, "revokeObjectURL");
});

describe("cached thumbnail lifecycle", () => {
  it("cleans both permission-renewal and expiry timers after unmount", async () => {
    jest.useFakeTimers();
    const scheduled = jest.spyOn(globalThis, "setTimeout");
    const cleared = jest.spyOn(globalThis, "clearTimeout");
    try {
      const view = render(
        <CachedAttachmentThumbnail fileName="Guide.pdf" url="/thumbnail" />,
      );
      await act(async () => undefined);
      const ownedTimers = scheduled.mock.calls.flatMap((call, index) =>
        call[1] === 25_000 || call[1] === 30_000
          ? [scheduled.mock.results[index].value]
          : [],
      );
      expect(ownedTimers).toHaveLength(2);
      view.unmount();
      ownedTimers.forEach((timer) => {
        expect(cleared).toHaveBeenCalledWith(timer);
      });
      await act(async () => {
        await jest.advanceTimersByTimeAsync(60_000);
      });
      expect(mockLoad).toHaveBeenCalledTimes(1);
    } finally {
      scheduled.mockRestore();
      cleared.mockRestore();
      jest.useRealTimers();
    }
  });
  it("paints a small cached image and releases its object URL on unmount", async () => {
    const view = render(
      <CachedAttachmentThumbnail fileName="Guide.pdf" url="/thumbnail" />,
    );
    await waitFor(() => {
      expect(screen.getByRole("img")).toHaveAttribute("aria-busy", "false");
    });
    expect(view.container.querySelector("img")).toHaveAttribute(
      "src",
      "blob:preview",
    );
    view.unmount();
    // eslint-disable-next-line @typescript-eslint/unbound-method -- Replaced with a Jest spy in setup.
    expect(URL.revokeObjectURL).toHaveBeenCalledWith("blob:preview");
    expect((mockLoad.mock.calls[0][1] as AbortSignal).aborted).toBe(true);
  });
  it("reports a failed derivative without blocking its parent's full viewer", async () => {
    mockLoad.mockRejectedValueOnce(new Error("Access denied"));
    const open = jest.fn();
    render(
      <button onClick={open} type="button">
        <CachedAttachmentThumbnail fileName="Guide.pdf" url="/thumbnail" />
      </button>,
    );
    expect(
      await screen.findByText("Preview unavailable · Open to view"),
    ).toBeVisible();
    fireEvent.click(screen.getByRole("button"));
    expect(open).toHaveBeenCalledTimes(1);
  });
  it("handles a browser decode failure with the same usable fallback", async () => {
    const view = render(
      <CachedAttachmentThumbnail fileName="Guide.pdf" url="/thumbnail" />,
    );
    await waitFor(() => {
      expect(view.container.querySelector("img")).not.toBeNull();
    });
    fireEvent.error(view.container.querySelector("img")!);
    expect(
      screen.getByText("Preview unavailable · Open to view"),
    ).toBeVisible();
  });
  it("passes explicit refresh intent and aborts a pending fetch on unmount", async () => {
    mockLoad.mockReturnValue(
      new Promise(() => {
        /* intentionally pending until unmount */
      }),
    );
    const view = render(
      <CachedAttachmentThumbnail
        fileName="Guide.pdf"
        refreshKey={1}
        url="/thumbnail"
      />,
    );
    await act(async () => undefined);
    expect(mockLoad.mock.calls[0][2]).toBe(true);
    view.unmount();
    expect((mockLoad.mock.calls[0][1] as AbortSignal).aborted).toBe(true);
  });
});
