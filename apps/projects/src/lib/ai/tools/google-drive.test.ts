/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { ApiError, get } from "@/lib/http";
import {
  redactGoogleDriveContentFromMessages,
  toGoogleDriveContentModelOutput,
} from "@/lib/ai/google-drive-tool-output";
import {
  getLinkedGoogleFileContentTool,
  listLinkedGoogleFilesTool,
} from "./google-drive";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  ApiContractError: class extends Error {},
  ApiError: class extends Error {
    status: number;

    constructor(message: string, status: number) {
      super(message);
      this.status = status;
    }
  },
  get: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getMock = jest.mocked(get);
const selectedFile = {
  referenceId: "00000000-0000-4000-8000-000000000001",
  name: "Launch plan",
  mimeType: "application/vnd.google-apps.document",
};
const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

const toolOptions: ToolExecutionOptions = {
  toolCallId: "tool-call-1",
  messages: [],
  experimental_context: {
    workspaceSlug: "complexus",
    selectedGoogleDriveFiles: [selectedFile],
  },
};

const executeTool = async <Input, Output>(
  execute: ToolExecuteFunction<Input, Output> | undefined,
  input: Input,
  options: ToolExecutionOptions = toolOptions,
): Promise<Output> => {
  if (!execute) throw new Error("Tool does not have an execute function");

  const result = execute(input, options);
  if (
    typeof result === "object" &&
    result !== null &&
    Symbol.asyncIterator in result
  ) {
    throw new Error("Streaming tool results are not supported by this test");
  }

  return (await result) as Output;
};

describe("Maya Google Drive tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("lists only files selected in the server-provided tool context", async () => {
    await expect(
      executeTool(listLinkedGoogleFilesTool.execute, {}),
    ).resolves.toEqual({
      success: true,
      count: 1,
      files: [selectedFile],
    });
    expect(getMock).not.toHaveBeenCalled();
  });

  it("refuses an arbitrary valid UUID that was not explicitly selected", async () => {
    await expect(
      executeTool(getLinkedGoogleFileContentTool.execute, {
        referenceId: "00000000-0000-4000-8000-000000000002",
      }),
    ).resolves.toEqual({
      success: false,
      error:
        "That Google Drive file was not explicitly selected for this request.",
    });
    expect(getMock).not.toHaveBeenCalled();
  });

  it("reads through the actor-scoped backend endpoint and bounds content", async () => {
    getMock.mockImplementation(async (_path, _ctx, _options, decode) => {
      const value = {
        data: {
          referenceId: selectedFile.referenceId,
          name: selectedFile.name,
          mimeType: selectedFile.mimeType,
          webViewLink: "https://docs.google.com/document/d/example/edit",
          modifiedTime: "2026-09-03T10:00:00Z",
          content: "x".repeat(100_100),
          contentType: "text/plain",
          truncated: false,
          bytesRead: 100_100,
        },
      };
      return decode ? decode(value) : value;
    });

    const result = await executeTool(getLinkedGoogleFileContentTool.execute, {
      referenceId: selectedFile.referenceId,
    });
    expect(getLinkedGoogleFileContentTool.toModelOutput).toBe(
      toGoogleDriveContentModelOutput,
    );
    const modelOutput = toGoogleDriveContentModelOutput({ output: result });

    expect(getMock).toHaveBeenCalledWith(
      `google-drive/files/${selectedFile.referenceId}/content`,
      expect.objectContaining({
        session,
        workspaceSlug: "complexus",
      }),
      { timeout: 20_000 },
      expect.any(Function),
    );
    expect(result).toHaveProperty("file.contentRetained", false);
    expect(result).toHaveProperty("file.contentTruncated", true);
    expect(JSON.stringify(result)).not.toContain("x".repeat(100));
    expect(modelOutput).toHaveProperty("value.file.content");
    expect(modelOutput).toHaveProperty(
      "value.file.contentAvailableForCurrentResponse",
      true,
    );
    expect(
      (
        modelOutput as unknown as {
          value: { file: { content: string } };
        }
      ).value.file.content,
    ).toHaveLength(20_000);

    const replayedOutput = JSON.parse(JSON.stringify(result)) as unknown;
    const replayedModelOutput = toGoogleDriveContentModelOutput({
      output: replayedOutput,
    });
    expect(replayedModelOutput).not.toHaveProperty("value.file.content");
    expect(replayedModelOutput).toHaveProperty(
      "value.file.contentAvailableForCurrentResponse",
      false,
    );
  });

  it("redacts legacy Drive content before transcript persistence", () => {
    const content = "Private launch plan that must not be retained";
    const messages = redactGoogleDriveContentFromMessages([
      {
        id: "assistant-1",
        role: "assistant",
        parts: [
          {
            type: "tool-getLinkedGoogleFileContentTool",
            state: "output-available",
            toolCallId: "tool-call-1",
            input: { referenceId: selectedFile.referenceId },
            output: {
              success: true,
              file: {
                ...selectedFile,
                webViewLink: "https://docs.google.com/document/d/example/edit",
                content,
                contentType: "text/plain",
                contentTruncated: false,
                bytesRead: content.length,
                untrustedExternalContent: true,
              },
            },
          },
        ],
      },
    ]);

    expect(JSON.stringify(messages)).not.toContain(content);
    expect(messages).toHaveProperty(
      "0.parts.0.output.file.contentRetained",
      false,
    );
  });

  it("does not expose provider error payloads to the model", async () => {
    getMock.mockRejectedValue(
      new ApiError("provider body containing a private token", 403, {
        token: "private token",
      }),
    );

    const result = await executeTool(getLinkedGoogleFileContentTool.execute, {
      referenceId: selectedFile.referenceId,
    });

    expect(result).toEqual({
      success: false,
      error: "You no longer have access to this selected Google Drive file.",
    });
    expect(JSON.stringify(result)).not.toContain("private token");
  });

  it("rejects an untrusted file link returned by the backend contract", async () => {
    getMock.mockImplementation(async (_path, _ctx, _options, decode) => {
      const value = {
        data: {
          referenceId: selectedFile.referenceId,
          name: selectedFile.name,
          mimeType: selectedFile.mimeType,
          webViewLink: "https://evil.example.test/steal",
          content: "Private plan",
          contentType: "text/plain",
          truncated: false,
          bytesRead: 12,
        },
      };
      return decode ? decode(value) : value;
    });

    await expect(
      executeTool(getLinkedGoogleFileContentTool.execute, {
        referenceId: selectedFile.referenceId,
      }),
    ).resolves.toEqual({
      success: false,
      error: "Google Drive returned an invalid content response.",
    });
  });
});
