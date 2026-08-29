/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, cleanup, render, screen } from "@testing-library/react";
import {
  KeyResultDeepLinkTarget,
  resolveTargetKeyResultId,
} from "./key-result-deep-link";

const scrollIntoView = jest.fn();

describe("key-result deep links", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    scrollIntoView.mockClear();
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
      configurable: true,
      value: scrollIntoView,
    });
    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      value: jest.fn((query: string) => ({
        addEventListener: jest.fn(),
        matches: query === "(min-width: 768px)",
        media: query,
        removeEventListener: jest.fn(),
      })),
    });
    Object.defineProperty(window, "requestAnimationFrame", {
      configurable: true,
      value: jest.fn((callback: FrameRequestCallback) => {
        callback(0);
        return 1;
      }),
    });
    Object.defineProperty(window, "cancelAnimationFrame", {
      configurable: true,
      value: jest.fn(),
    });
  });

  afterEach(() => {
    cleanup();
    jest.clearAllTimers();
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  it("only accepts a requested key result that belongs to the loaded set", () => {
    expect(
      resolveTargetKeyResultId("key-result-2", [
        "key-result-1",
        "key-result-2",
      ]),
    ).toBe("key-result-2");
    expect(
      resolveTargetKeyResultId("another-objective-key-result", [
        "key-result-1",
      ]),
    ).toBeNull();
  });

  it("scrolls, focuses, and briefly highlights the active viewport row", () => {
    render(
      <KeyResultDeepLinkTarget
        id="key-result-1"
        isTarget
        name="Increase activation"
        viewport="desktop"
      >
        Row content
      </KeyResultDeepLinkTarget>,
    );

    const row = screen.getByRole("group", {
      name: "Key result: Increase activation",
    });
    expect(scrollIntoView).toHaveBeenCalledWith({
      behavior: "smooth",
      block: "center",
      inline: "nearest",
    });
    expect(row).toHaveFocus();
    expect(row).toHaveClass("ring-1");

    act(() => {
      jest.runOnlyPendingTimers();
    });
    expect(row).not.toHaveClass("ring-1");
  });

  it("does not focus the hidden viewport copy", () => {
    render(
      <KeyResultDeepLinkTarget
        id="key-result-1"
        isTarget
        name="Increase activation"
        viewport="mobile"
      >
        Row content
      </KeyResultDeepLinkTarget>,
    );

    expect(scrollIntoView).not.toHaveBeenCalled();
    expect(document.body).toHaveFocus();
  });
});
