/* global afterEach, describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import {
  getRequestOptionsScope,
  installRequestOptionsScopeResolver,
} from "./request-options-scope";

const cleanups: (() => void)[] = [];

afterEach(() => {
  while (cleanups.length > 0) cleanups.pop()?.();
});

describe("request options scope", () => {
  it("exposes the active scope without importing server-only modules", () => {
    const signal = new AbortController().signal;
    cleanups.push(installRequestOptionsScopeResolver(() => ({ signal })));

    expect(getRequestOptionsScope()).toEqual({ signal });

    const source = readFileSync(
      __filename.replace(/\.test\.ts$/, ".ts"),
      "utf8",
    );
    expect(source).not.toMatch(/(?:node:async_hooks|server-only)/);
  });

  it("restores a previous resolver without clobbering a newer install", () => {
    const firstSignal = new AbortController().signal;
    const secondSignal = new AbortController().signal;
    const uninstallFirst = installRequestOptionsScopeResolver(() => ({
      signal: firstSignal,
    }));
    const uninstallSecond = installRequestOptionsScopeResolver(() => ({
      signal: secondSignal,
    }));

    uninstallFirst();
    expect(getRequestOptionsScope()?.signal).toBe(secondSignal);

    uninstallSecond();
    expect(getRequestOptionsScope()?.signal).toBe(firstSignal);
    uninstallFirst();
  });
});
