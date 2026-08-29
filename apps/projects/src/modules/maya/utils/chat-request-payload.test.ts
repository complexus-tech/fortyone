/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  HISTORICAL_ATTACHMENT_PLACEHOLDER,
  omitHistoricalFileParts,
  prepareMayaChatSendRequest,
} from "./chat-request-payload";

const filePart = (filename: string, url: string) => ({
  filename,
  mediaType: "text/plain",
  type: "file" as const,
  url,
});

const textPart = (text: string) => ({ text, type: "text" as const });

describe("Maya chat request payload", () => {
  it("replaces file parts outside the last user turn with a fixed placeholder", () => {
    const messages: MayaUIMessage[] = [
      {
        id: "old-user",
        role: "user",
        parts: [
          textPart("Review the old attachment."),
          filePart("private-old.txt", "data:text/plain;base64,b2xk"),
        ],
      },
      {
        id: "assistant-file",
        role: "assistant",
        parts: [
          filePart("generated-old.txt", "data:text/plain;base64,b2xkMg=="),
        ],
      },
      {
        id: "current-user",
        role: "user",
        parts: [
          textPart("Compare this one."),
          filePart("current.txt", "data:text/plain;base64,Y3VycmVudA=="),
        ],
      },
    ];

    const preparedMessages = omitHistoricalFileParts(messages);
    const serialized = JSON.stringify(preparedMessages);

    expect(preparedMessages[0]?.parts).toEqual([
      textPart("Review the old attachment."),
      textPart(HISTORICAL_ATTACHMENT_PLACEHOLDER),
    ]);
    expect(preparedMessages[1]?.parts).toEqual([
      textPart(HISTORICAL_ATTACHMENT_PLACEHOLDER),
    ]);
    expect(preparedMessages[2]).toBe(messages[2]);
    expect(serialized).not.toContain("private-old.txt");
    expect(serialized).not.toContain("generated-old.txt");
    expect(serialized).toContain("current.txt");
    expect(serialized).toContain("data:text/plain;base64,Y3VycmVudA==");
  });

  it("removes every file when the request has no user turn", () => {
    const messages: MayaUIMessage[] = [
      {
        id: "assistant-file",
        role: "assistant",
        parts: [filePart("result.txt", "data:text/plain;base64,cmVzdWx0")],
      },
    ];

    expect(omitHistoricalFileParts(messages)[0]?.parts).toEqual([
      textPart(HISTORICAL_ATTACHMENT_PLACEHOLDER),
    ]);
  });

  it("builds the complete transport body with compacted messages", () => {
    const messages: MayaUIMessage[] = [
      {
        id: "old-user",
        role: "user",
        parts: [filePart("old.txt", "data:text/plain;base64,b2xk")],
      },
      {
        id: "current-user",
        role: "user",
        parts: [textPart("Continue without another attachment.")],
      },
    ];

    const request = prepareMayaChatSendRequest({
      api: "/api/chat",
      body: { workspace: { slug: "first" } },
      credentials: undefined,
      headers: undefined,
      id: "chat-1",
      messageId: "current-user",
      messages,
      requestMetadata: undefined,
      trigger: "submit-message",
    });

    expect(request.body).toMatchObject({
      id: "chat-1",
      messageId: "current-user",
      trigger: "submit-message",
      workspace: { slug: "first" },
    });
    expect(request.body.messages[0]?.parts).toEqual([
      textPart(HISTORICAL_ATTACHMENT_PLACEHOLDER),
    ]);
  });

  it("does not resend the last user attachment for an approval auto-submit", () => {
    const messages: MayaUIMessage[] = [
      {
        id: "user-with-file",
        role: "user",
        parts: [filePart("brief.pdf", "data:application/pdf;base64,cGRm")],
      },
      {
        id: "assistant-approval",
        role: "assistant",
        parts: [textPart("Prepared for approval.")],
      },
    ];

    const request = prepareMayaChatSendRequest({
      api: "/api/chat",
      body: undefined,
      credentials: undefined,
      headers: undefined,
      id: "chat-1",
      messageId: "assistant-approval",
      messages,
      requestMetadata: undefined,
      trigger: "submit-message",
    });

    expect(JSON.stringify(request.body.messages)).not.toContain("cGRm");
    expect(request.body.messages[0]?.parts).toEqual([
      textPart(HISTORICAL_ATTACHMENT_PLACEHOLDER),
    ]);
  });
});
