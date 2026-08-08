/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const source = readFileSync(
  join(process.cwd(), "src/modules/documents/documents-home.tsx"),
  "utf8",
);

describe("Documents home controls", () => {
  it("groups list filters in one popover", () => {
    expect(source).toContain("const DocumentFilters");
    expect(source).toContain('<Popover.Content align="start"');
    expect(source).toContain('label="Access"');
    expect(source).toContain('label="Owner"');
    expect(source).toContain('label="Updated"');
    expect(source).not.toContain("DocumentControlMenu");
  });

  it("uses compact sort labels and icon pagination in the top toolbar", () => {
    expect(source).toContain('label: "Newest"');
    expect(source).toContain('label: "Oldest"');
    expect(source).toContain('label: "A to Z"');
    expect(source).toContain('label: "Z to A"');
    expect(source).toContain('aria-label="Previous page"');
    expect(source).toContain('aria-label="Next page"');
    expect(source).toContain(
      "Page {pagination.page} of {pagination.pageCount}",
    );
    expect(source).toContain('className="h-5 w-auto"');
    expect(source).not.toContain("{filteredDocuments.length}");
  });
});
