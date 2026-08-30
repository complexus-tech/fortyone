/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getRequestAttachments,
  getRequestExternalLinks,
} from "./request-metadata";

describe("integration request metadata", () => {
  it("normalizes supported external link shapes and ignores invalid values", () => {
    expect(
      getRequestExternalLinks({
        links: [
          "https://linear.app/example",
          { href: "https://github.com/example", label: "GitHub" },
          { title: "Missing URL" },
          null,
        ],
      }),
    ).toEqual([
      {
        title: "https://linear.app/example",
        url: "https://linear.app/example",
      },
      { title: "GitHub", url: "https://github.com/example" },
    ]);
  });

  it("keeps attachment labels when a source does not provide a URL", () => {
    expect(
      getRequestAttachments({
        attachments: [
          { filename: "notes.md" },
          { name: "Brief", url: "https://example.com/brief" },
          { name: "" },
        ],
      }),
    ).toEqual([
      { name: "notes.md", url: undefined },
      { name: "Brief", url: "https://example.com/brief" },
    ]);
  });
});
