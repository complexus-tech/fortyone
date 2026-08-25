/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- The AI SDK requires web streams.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { tool } from "ai";
import { z } from "zod";
import { withCompactModelOutputs } from "./model-tools";

describe("withCompactModelOutputs", () => {
  it("keeps raw execution output while compacting the model-facing result", async () => {
    const rawOutput = {
      description: "Visible to the UI",
      descriptionHTML: `<p>${"large".repeat(500)}</p>`,
      success: true,
    };
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object({}),
        execute: () => rawOutput,
      }),
    });

    expect(toolSet.example.execute?.({}, {} as never)).toEqual(rawOutput);
    expect(
      await toolSet.example.toModelOutput?.({
        input: {},
        output: rawOutput,
        toolCallId: "tool-call-1",
      }),
    ).toEqual({
      type: "json",
      value: {
        description: "Visible to the UI",
        success: true,
      },
    });
  });
});
