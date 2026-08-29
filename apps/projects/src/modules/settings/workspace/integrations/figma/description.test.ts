/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  figmaDescriptionRequestSchema,
  figmaDescriptionSchema,
} from "./description";

describe("Figma description schemas", () => {
  it("accepts a concise structured story description", () => {
    expect(
      figmaDescriptionSchema.safeParse({
        overview: "Let customers review their order before payment.",
        requirements: ["Show the order total"],
        acceptanceCriteria: ["The customer can review the order total"],
        implementationNotes: [],
      }).success,
    ).toBe(true);
  });

  it("rejects requests without extracted design text", () => {
    expect(
      figmaDescriptionRequestSchema.safeParse({
        fileName: "Checkout",
        nodeName: "Review order",
        nodeType: "FRAME",
        textContent: [],
      }).success,
    ).toBe(false);
  });
});
