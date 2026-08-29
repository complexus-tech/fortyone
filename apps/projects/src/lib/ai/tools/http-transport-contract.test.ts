/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

const TOOL_DIRECTORY = __dirname;

const getSourceFiles = (directory: string): string[] =>
  readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) return getSourceFiles(path);
    if (
      !/\.(?:js|jsx|ts|tsx)$/.test(entry.name) ||
      /\.(?:test|spec)\.(?:js|jsx|ts|tsx)$/.test(entry.name)
    ) {
      return [];
    }
    return [path];
  });

const toolSourceFiles = getSourceFiles(TOOL_DIRECTORY);

describe("Maya tool HTTP transport contract", () => {
  it.each(toolSourceFiles)(
    "%s does not bypass the scoped workspace HTTP transport",
    (path) => {
      const source = readFileSync(path, "utf8");

      expect(source).not.toMatch(/\b(?:(?:globalThis|window)\.)?fetch\s*\(/);
      expect(source).not.toMatch(
        /(?:\bfrom\s*|\bimport\s*(?:\(\s*)?|\brequire\s*\(\s*)["']api-client(?:\/[^"']*)?["']/,
      );
    },
  );
});
