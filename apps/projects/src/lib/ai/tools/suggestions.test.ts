/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- The AI SDK requires web streams.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { asSchema } from "ai";
import { suggestions } from "./suggestions";

describe("suggestions tool", () => {
  it("accepts only the documented two or three follow-up actions", async () => {
    const schema = asSchema(suggestions.inputSchema);

    await expect(
      schema.validate?.({ suggestions: ["Open it", "Assign it"] }),
    ).resolves.toEqual({
      success: true,
      value: { suggestions: ["Open it", "Assign it"] },
    });
    await expect(
      schema.validate?.({
        suggestions: ["One", "Two", "Three", "Four"],
      }),
    ).resolves.toEqual(expect.objectContaining({ success: false }));
  });
});
