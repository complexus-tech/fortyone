/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  formatChatErrorDiagnostic,
  getChatErrorDiagnostic,
  getChatStreamErrorMessage,
} from "./chat-errors";

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

  it("finds a rate limit nested inside an SDK retry error", () => {
    const message = getChatStreamErrorMessage({
      lastError: {
        cause: {
          error: {
            code: "rate_limit_exceeded",
            message: "Too many requests",
          },
        },
      },
      message: "Failed after 2 attempts",
    });

    expect(message).toContain("temporary AI rate limit");
  });

  it("returns mutation-safe guidance for timeouts", () => {
    const message = getChatStreamErrorMessage({
      errors: [new Error("request timed out while reading the stream")],
    });

    expect(message).toContain("took too long");
    expect(message).toContain("verify the result");
  });

  it("does not expose unknown provider errors", () => {
    expect(
      getChatStreamErrorMessage(
        new Error("upstream detail with internal identifiers"),
      ),
    ).toBe("Maya couldn't finish the response. Please try again in a moment.");
  });

  it("gives actionable guidance for an oversized request", () => {
    expect(
      getChatStreamErrorMessage({
        code: "request_too_large",
        message: "Maya chat request is too large.",
      }),
    ).toContain("Remove an attachment or split the request");
  });
});

describe("getChatErrorDiagnostic", () => {
  it("keeps payloads and user content out of server diagnostics", () => {
    const error = Object.assign(new Error("secret user request"), {
      code: "provider_error",
      requestBody: { messages: ["private workspace content"] },
      status: 400,
      cause: {
        code: "invalid_tool_input",
        value: { title: "confidential ticket" },
      },
    });

    const diagnostic = getChatErrorDiagnostic(error);

    expect(diagnostic).toEqual({
      codes: ["provider_error", "invalid_tool_input"],
      errorType: "Error",
      statuses: [400],
    });
    expect(JSON.stringify(diagnostic)).not.toMatch(
      /secret|private|confidential/i,
    );
  });

  it("captures allowlisted HTTP diagnostics without serializing response data", () => {
    const error = Object.assign(new Error("internal server error"), {
      data: {
        data: { transcript: "private workspace content" },
        error: {
          code: "internal_error",
          message: "internal server error",
          request_id: "request-123",
        },
      },
      name: "ApiError",
      status: 500,
    });

    const diagnostic = getChatErrorDiagnostic(error);

    expect(diagnostic).toEqual({
      codes: ["internal_error"],
      errorType: "ApiError",
      messages: ["internal server error"],
      requestIds: ["request-123"],
      statuses: [500],
    });
    expect(formatChatErrorDiagnostic(error)).toBe(JSON.stringify(diagnostic));
    expect(JSON.stringify(diagnostic)).not.toContain("private workspace content");
  });

  it("captures AI SDK statusCode and retryability fields", () => {
    const error = Object.assign(new Error("Provider request failed"), {
      isRetryable: false,
      name: "AI_APICallError",
      requestBodyValues: { messages: ["private prompt"] },
      statusCode: 400,
    });

    expect(getChatErrorDiagnostic(error)).toEqual({
      codes: [],
      errorType: "AI_APICallError",
      messages: ["Provider request failed"],
      retryable: [false],
      statuses: [400],
    });
  });

  it("does not log validation messages that can embed private input values", () => {
    const error = Object.assign(
      new Error(
        'Type validation failed: Value: {"prompt":"private workspace content"}.',
      ),
      {
        name: "AI_TypeValidationError",
        statusCode: 400,
      },
    );

    const diagnostic = getChatErrorDiagnostic(error);

    expect(diagnostic).toEqual({
      codes: [],
      errorType: "AI_TypeValidationError",
      statuses: [400],
    });
    expect(JSON.stringify(diagnostic)).not.toContain("private workspace content");
  });
});
