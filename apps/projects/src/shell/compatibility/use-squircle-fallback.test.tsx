/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { StrictMode, Suspense, use } from "react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import { useSquircleFallback } from "./use-squircle-fallback";

const mockDestroy = jest.fn();
const mockRegister = jest.fn();

jest.mock("hyperellipse", () => ({
  registerHyperellipse: (...args: unknown[]) => mockRegister(...args),
}));

describe("useSquircleFallback", () => {
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

  it("enables the fallback in unsupported browsers without a desktop bridge", async () => {
    const { rerender } = renderHook(useSquircleFallback);
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledTimes(1);
    });
    expect(CSS.supports).toHaveBeenCalledWith("corner-shape", "squircle");
    expect(mockRegister).toHaveBeenCalledWith({
      pendingRadiusScale: 0.75 / 2.2,
    });
    expect(document.documentElement.dataset.squircleFallback).toBe("true");

    rerender();
    expect(mockRegister).toHaveBeenCalledTimes(1);
  });

  it("keeps native corner rendering when the browser supports it", async () => {
    jest.spyOn(CSS, "supports").mockReturnValue(true);
    renderHook(useSquircleFallback);
    await act(async () => {});
    expect(CSS.supports).toHaveBeenCalledWith("corner-shape", "squircle");
    expect(mockRegister).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("retains ordinary styles when feature detection is unavailable", async () => {
    Object.defineProperty(window, "CSS", {
      configurable: true,
      value: undefined,
    });
    renderHook(useSquircleFallback);
    await act(async () => {});
    expect(mockRegister).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("removes scoped styles and the renderer on teardown", async () => {
    const { unmount } = renderHook(useSquircleFallback);
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledTimes(1);
    });
    expect(document.documentElement.dataset.squircleFallback).toBe("true");
    unmount();
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("does not start a renderer after its provider unmounts", async () => {
    const { unmount } = renderHook(useSquircleFallback);
    unmount();
    await act(async () => {});
    expect(mockRegister).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("keeps one active renderer through Strict Mode effect replay", async () => {
    const { unmount } = renderHook(useSquircleFallback, {
      wrapper: StrictMode,
    });
    await waitFor(() => {
      expect(mockRegister).toHaveBeenCalledTimes(1);
    });
    expect(mockDestroy).not.toHaveBeenCalled();
    expect(document.documentElement.dataset.squircleFallback).toBe("true");

    unmount();
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("waits for a suspended workspace shell to hydrate before painting", async () => {
    let suspendHydration = false;
    let resumeHydration: (() => void) | undefined;
    const hydration = new Promise<void>((resolve) => {
      resumeHydration = resolve;
    });
    const HydrationGate = ({ children }: { children: ReactNode }) => {
      if (suspendHydration) use(hydration);
      return children;
    };
    const WorkspaceShell = () => {
      useSquircleFallback();
      return (
        <button style={{ borderRadius: "12px" }} type="button">
          Create story
        </button>
      );
    };
    const tree = (
      <Suspense fallback={<div>Loading workspace</div>}>
        <HydrationGate>
          <WorkspaceShell />
        </HydrationGate>
      </Suspense>
    );
    const container = document.createElement("div");
    container.innerHTML = renderToString(tree);
    const serverMarkup = container.innerHTML;
    document.body.append(container);
    suspendHydration = true;
    const onRecoverableError = jest.fn();
    const root = hydrateRoot(container, tree, { onRecoverableError });

    try {
      await act(async () => {});
      expect(mockRegister).not.toHaveBeenCalled();
      expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
      expect(container.innerHTML).toBe(serverMarkup);

      await act(async () => {
        suspendHydration = false;
        resumeHydration?.();
      });
      await waitFor(() => {
        expect(mockRegister).toHaveBeenCalledTimes(1);
      });
      expect(document.documentElement.dataset.squircleFallback).toBe("true");
      expect(onRecoverableError).not.toHaveBeenCalled();
    } finally {
      await act(async () => {
        root.unmount();
      });
      container.remove();
    }
    expect(mockDestroy).toHaveBeenCalledTimes(1);
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });

  it("restores the ordinary corner fallback if initialization fails", async () => {
    const error = new Error("Renderer unavailable");
    mockRegister.mockImplementation(() => {
      throw error;
    });
    const log = jest.spyOn(console, "error").mockImplementation(() => {});
    renderHook(useSquircleFallback);
    await waitFor(() => {
      expect(log).toHaveBeenCalledWith(
        "Could not initialize squircle fallback",
        error,
      );
    });
    expect(document.documentElement.dataset.squircleFallback).toBeUndefined();
  });
});
