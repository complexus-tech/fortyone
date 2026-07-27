/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getPromptTextSegments } from "./chat-message-utils";

describe("getPromptTextSegments", () => {
  it("detects HTTP and www links without including sentence punctuation", () => {
    expect(
      getPromptTextSegments(
        "Review https://fortyone.app/work/PRD-571, then visit www.example.com.",
      ),
    ).toEqual([
      { start: 0, type: "text", value: "Review " },
      {
        type: "link",
        href: "https://fortyone.app/work/PRD-571",
        start: 7,
        value: "https://fortyone.app/work/PRD-571",
      },
      { start: 40, type: "text", value: "," },
      { start: 41, type: "text", value: " then visit " },
      {
        type: "link",
        href: "https://www.example.com",
        start: 53,
        value: "www.example.com",
      },
      { start: 68, type: "text", value: "." },
    ]);
  });

  it("leaves text without links unchanged", () => {
    expect(getPromptTextSegments("Plan the next sprint")).toEqual([
      { start: 0, type: "text", value: "Plan the next sprint" },
    ]);
  });
});
