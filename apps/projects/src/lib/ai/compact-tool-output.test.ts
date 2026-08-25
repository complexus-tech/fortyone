/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { compactToolOutput } from "./compact-tool-output";

describe("compactToolOutput", () => {
  it("removes UI-only fields and bounds arrays for model context", () => {
    const output = compactToolOutput({
      description: "Useful text",
      descriptionHTML: "<p>Useful text</p>",
      imageUrl: "https://example.com/image.png",
      items: Array.from({ length: 30 }, (_, index) => ({ id: index })),
      success: true,
    }) as Record<string, unknown>;

    expect(output.description).toBe("Useful text");
    expect(output).not.toHaveProperty("descriptionHTML");
    expect(output).not.toHaveProperty("imageUrl");
    expect(output.items).toHaveLength(20);
    expect(output.modelItemsOmitted).toEqual({ items: 10 });
  });

  it("truncates long strings and removes encoded binary payloads", () => {
    const output = compactToolOutput({
      attachment: `data:image/png;base64,${"a".repeat(2000)}`,
      content: "x".repeat(2000),
    }) as Record<string, string>;

    expect(output.attachment).toBe("[binary data omitted]");
    expect(output.content.length).toBeLessThan(1300);
    expect(output.content.endsWith("…")).toBe(true);
  });
});
