/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getChatStreamErrorMessage } from "./chat-errors";

describe("getChatStreamErrorMessage", () => {
  it("replaces a streamed rate-limit payload with an actionable message", () => {
    const message = getChatStreamErrorMessage({
      error: {
        code: "rate_limit_exceeded",
        message: "Rate limit reached for gpt-5.6-luna",
      },
      sequence_number: 2,
      type: "error",
    });

    expect(message).toContain("temporary AI rate limit");
    expect(message).toContain("do not repeat it");
    expect(message).not.toContain("gpt-5.6-luna");
  });

  it("handles the JSON string emitted by the streaming SDK", () => {
    const message = getChatStreamErrorMessage(
      new Error(
        JSON.stringify({
          error: {
            code: "rate_limit_exceeded",
            message: "Rate limit reached",
          },
          type: "error",
        }),
      ),
    );

    expect(message).toContain("temporary AI rate limit");
  });

  it("does not expose unknown provider errors", () => {
    expect(
      getChatStreamErrorMessage(
        new Error("upstream detail with internal identifiers"),
      ),
    ).toBe("Maya couldn't finish the response. Please try again in a moment.");
  });
});
