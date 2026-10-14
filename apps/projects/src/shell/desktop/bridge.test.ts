/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { readDesktopBackground } from "./bridge";

describe("readDesktopBackground", () => {
  afterEach(() => {
    document.documentElement.style.removeProperty("--color-background");
    document.body.style.removeProperty("background-color");
    jest.restoreAllMocks();
  });

  it("resolves the theme independently of a transparent body across theme changes", () => {
    const context = {
      fillStyle: "",
      fillRect: jest.fn(),
      getImageData: jest.fn().mockReturnValue({
        data: new Uint8ClampedArray([251, 250, 248, 255]),
      }),
    };
    jest
      .spyOn(HTMLCanvasElement.prototype, "getContext")
      .mockReturnValue(context as unknown as CanvasRenderingContext2D);
    document.body.style.backgroundColor = "transparent";
    document.documentElement.style.setProperty(
      "--color-background",
      "oklch(0.997 0.002 95)",
    );

    expect(readDesktopBackground()).toEqual([251, 250, 248]);
    expect(context.fillStyle).toBe("oklch(0.997 0.002 95)");

    document.documentElement.style.setProperty(
      "--color-background",
      "oklch(0.1821 0 0)",
    );
    context.getImageData.mockReturnValue({
      data: new Uint8ClampedArray([18, 18, 18, 255]),
    });
    expect(readDesktopBackground()).toEqual([18, 18, 18]);
    expect(context.fillStyle).toBe("oklch(0.1821 0 0)");
  });

  it("waits for theme styles before synchronizing native chrome", () => {
    expect(readDesktopBackground()).toBeNull();
  });
});
