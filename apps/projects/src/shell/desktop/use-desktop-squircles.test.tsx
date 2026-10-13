/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { act, renderHook, waitFor } from "@testing-library/react";
import { useDesktopSquircles } from "./use-desktop-squircles";

const mockDestroy = jest.fn();
const mockRegister = jest.fn();

jest.mock("hyperellipse", () => ({
  registerHyperellipse: (...args: unknown[]) => mockRegister(...args),
}));

describe("useDesktopSquircles", () => {
  const originalCSS = window.CSS;

  beforeEach(() => {
    mockDestroy.mockClear();
    mockRegister.mockReset().mockReturnValue({
      active: true,
      supported: false,
      destroy: mockDestroy,
      refresh: jest.fn(),
    });
    Object.defineProperty(window, "CSS", {
      configurable: true,
      value: { supports: jest.fn().mockReturnValue(false) },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, "CSS", {
      configurable: true,
      value: originalCSS,
    });
    jest.restoreAllMocks();
  });

  it("leaves ordinary browsers on their existing styles", async () => {
    renderHook(() => {
      useDesktopSquircles(false);
    });
    await act(async () => {});
    expect(mockRegister).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.desktopSquircles).toBeUndefined();
  });

  it("keeps native corner rendering when the desktop webview supports it", async () => {
    jest.spyOn(CSS, "supports").mockReturnValue(true);
    renderHook(() => {
      useDesktopSquircles(true);
    });
    await act(async () => {});
    expect(CSS.supports).toHaveBeenCalledWith("corner-shape", "squircle");
    expect(mockRegister).not.toHaveBeenCalled();
  });

  it("enables scoped desktop styles and removes the renderer on teardown", async () => {
    const { unmount } = renderHook(() => {
      useDesktopSquircles(true);
    });
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledTimes(1);
    });
    expect(document.documentElement.dataset.desktopSquircles).toBe("true");
    unmount();
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(document.documentElement.dataset.desktopSquircles).toBeUndefined();
  });

  it("does not start a renderer after its provider unmounts", async () => {
    const { unmount } = renderHook(() => {
      useDesktopSquircles(true);
    });
    unmount();
    await act(async () => {});
    expect(mockRegister).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.desktopSquircles).toBeUndefined();
  });

  it("restores the ordinary corner fallback if initialization fails", async () => {
    const error = new Error("Renderer unavailable");
    mockRegister.mockImplementation(() => {
      throw error;
    });
    const log = jest.spyOn(console, "error").mockImplementation(() => {});
    renderHook(() => {
      useDesktopSquircles(true);
    });
    await waitFor(() => {
      expect(log).toHaveBeenCalledWith(
        "Could not initialize desktop squircles",
        error,
      );
    });
    expect(document.documentElement.dataset.desktopSquircles).toBeUndefined();
  });
});
