/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { UIMessage } from "ai";
import {
  getChatTitle,
  getChatTitleSource,
  normalizeGeneratedChatTitle,
} from "./chat-title";

const userMessage = (text: string): UIMessage => ({
  id: "message-1",
  parts: [{ text, type: "text" }],
  role: "user",
});

describe("getChatTitle", () => {
  it("derives a short title locally from the first user message", () => {
    expect(
      getChatTitle([
        userMessage(
          "Bulk create onboarding stories for the Product team. Include acceptance criteria.",
        ),
      ]),
    ).toBe("Bulk create onboarding stories for the Product team.");
  });

  it("truncates long titles at a readable word boundary", () => {
    const title = getChatTitle([
      userMessage(
        "Please analyze every objective across every team and prepare a comprehensive delivery risk report for leadership",
      ),
    ]);

    expect(title.length).toBeLessThanOrEqual(64);
    expect(title.endsWith("…")).toBe(true);
  });

  it("uses a stable fallback when no text is available", () => {
    expect(
      getChatTitle([
        {
          id: "file-message",
          parts: [],
          role: "user",
        },
      ]),
    ).toBe("New conversation");
  });

  it("bounds the text sent to the title model", () => {
    expect(getChatTitleSource([userMessage("x".repeat(2000))])).toHaveLength(
      500,
    );
  });

  it("normalizes a generated title without trusting verbose output", () => {
    expect(
      normalizeGeneratedChatTitle('Title: "Plan Product Launch"\nNotes'),
    ).toBe("Plan Product Launch");
  });
});
