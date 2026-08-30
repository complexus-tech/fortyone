/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const documentsHomeSource = readFileSync(
  join(process.cwd(), "src/modules/documents/documents-home.tsx"),
  "utf8",
);

const controlsSource = readFileSync(
  join(process.cwd(), "src/modules/documents/documents-home-controls.tsx"),
  "utf8",
);

describe("Documents home controls", () => {
  it("groups list filters in one popover", () => {
    expect(controlsSource).toContain("DocumentFilters");
    expect(controlsSource).toContain('<Popover.Content align="start"');
    expect(controlsSource).toContain('label="Access"');
    expect(controlsSource).toContain('label="Owner"');
    expect(controlsSource).toContain('label="Updated"');
    expect(controlsSource).not.toContain("DocumentControlMenu");
  });

  it("uses compact sort labels and icon pagination in the top toolbar", () => {
    expect(controlsSource).toContain('label: "Newest"');
    expect(controlsSource).toContain('label: "Oldest"');
    expect(controlsSource).toContain('label: "A to Z"');
    expect(controlsSource).toContain('label: "Z to A"');
    expect(documentsHomeSource).toContain('aria-label="Previous page"');
    expect(documentsHomeSource).toContain('aria-label="Next page"');
    expect(documentsHomeSource).toContain(
      "Page {pagination.page} of {pagination.pageCount}",
    );
    expect(documentsHomeSource).toContain('className="h-5 w-auto"');
    expect(documentsHomeSource).not.toContain("{filteredDocuments.length}");
  });
});
