/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const source = readFileSync(
  join(process.cwd(), "src/modules/documents/document-page.tsx"),
  "utf8",
);
const styles = readFileSync(
  join(process.cwd(), "src/modules/documents/document-page.module.css"),
  "utf8",
);

describe("document page header", () => {
  it("layers a borderless gradient header over the scrolling document", () => {
    expect(source).toContain(
      "pointer-events-none absolute inset-x-0 top-0 z-20",
    );
    expect(source).toContain("styles.headerBackdrop");
    expect(source).toContain("top-0 z-20 h-18");
    expect(styles).toContain("background: linear-gradient(");
    expect(styles).not.toContain("backdrop-filter");
    expect(styles).not.toContain("mask-image");
    expect(styles).toContain("var(--document-canvas-color) 68%");
    expect(styles).toContain(":global(.dark) .headerBackdrop");
    expect(styles).toContain("var(--color-surface-muted) 50%");
    expect(source).not.toContain(
      'className="border-border/70 h-18 shrink-0 border-b',
    );
  });

  it("keeps the document title below the overlaid header at rest", () => {
    expect(source).toContain("pt-30 pb-32");
    expect(source).toContain("lg:pt-34");
  });

  it("keeps related work clear of the access controls", () => {
    expect(source).toContain('className="absolute top-24 right-5 z-30"');
    expect(source).toContain('rounded="full"');
    expect(source).toContain("Related work");
    expect(source).not.toContain("rounded-r-none border-r-0");
  });
});
