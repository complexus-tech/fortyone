/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const source = readFileSync(
  join(process.cwd(), "src/modules/documents/document-page.tsx"),
  "utf8",
);
const headerSource = readFileSync(
  join(process.cwd(), "src/modules/documents/document-page-header.tsx"),
  "utf8",
);
const styles = readFileSync(
  join(process.cwd(), "src/modules/documents/document-page.module.css"),
  "utf8",
);

describe("document page header", () => {
  it("layers a borderless gradient header over the scrolling document", () => {
    expect(headerSource).toContain(
      "pointer-events-none absolute inset-x-0 top-0 z-20",
    );
    expect(headerSource).toContain("styles.headerBackdrop");
    expect(headerSource).toContain("top-0 z-20 h-18");
    expect(styles).toContain("background: linear-gradient(");
    expect(styles).not.toContain("backdrop-filter");
    expect(styles).not.toContain("mask-image");
    expect(styles).toContain("var(--document-canvas-color) 68%");
    expect(styles).toContain(":global(.dark) .headerBackdrop");
    expect(styles).toContain("var(--color-surface-muted) 50%");
    expect(headerSource).not.toContain(
      'className="border-border/70 h-18 shrink-0 border-b',
    );
  });

  it("keeps the document title below the overlaid header at rest", () => {
    expect(source).toContain("pt-30 pb-32");
    expect(source).toContain("lg:pt-34");
  });

  it("places relationships inline with the document metadata", () => {
    expect(source).toContain("<DocumentRelationshipControl");
    expect(source).not.toContain('className="absolute top-24 right-5 z-30"');
    expect(source).not.toContain('aria-label="Show related work"');
  });
});
