/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getDocument, getDocuments } from "@/modules/documents/queries";
import type {
  WorkspaceDocument,
  WorkspaceDocumentSummary,
} from "@/modules/documents/types";
import { getDocumentDetailsTool, listDocumentsTool } from "./documents";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/documents/queries", () => ({
  getDocument: jest.fn(),
  getDocuments: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getDocumentMock = jest.mocked(getDocument);
const getDocumentsMock = jest.mocked(getDocuments);

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
  experimental_context: { workspaceSlug: "complexus" },
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

const summary = (id: string): WorkspaceDocumentSummary => ({
  id,
  workspaceId: "workspace-1",
  title: `Document ${id}`,
  visibility: "workspace",
  createdBy: "user-1",
  updatedBy: "user-1",
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-06T08:00:00.000Z",
  canEdit: true,
  relatedWorkCount: 1,
});

describe("Maya document tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("searches only documents visible through the authenticated query", async () => {
    getDocumentsMock.mockResolvedValue([
      summary("document-1"),
      summary("document-2"),
      summary("document-3"),
    ]);

    const result = await executeTool(listDocumentsTool.execute, {
      search: "  roadmap  ",
      scope: "shared" as const,
      limit: 2,
    });

    expect(getDocumentsMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      "roadmap",
      "shared",
      3,
    );
    expect(result).toMatchObject({
      success: true,
      count: 2,
      hasMore: true,
      filters: { scope: "shared", search: "roadmap" },
      documents: [
        { id: "document-1", title: "Document document-1" },
        { id: "document-2", title: "Document document-2" },
      ],
    });
  });

  it("returns full plain text and related work for an accessible document", async () => {
    const document: WorkspaceDocument = {
      ...summary("document-1"),
      contentHtml: "<p>Delivery plan</p>",
      contentText: "Delivery plan",
      sharedWith: [],
      relatedWork: [
        {
          entityId: "story-1",
          entityType: "story",
          title: "Ship the plan",
          reference: "WEB-12",
          teamId: "team-1",
        },
      ],
    };
    getDocumentMock.mockResolvedValue(document);

    const result = await executeTool(getDocumentDetailsTool.execute, {
      documentId: "document-1",
    });

    expect(getDocumentMock).toHaveBeenCalledWith("document-1", {
      session,
      workspaceSlug: "complexus",
    });
    expect(result).toMatchObject({
      success: true,
      document: {
        id: "document-1",
        content: "Delivery plan",
        contentTruncated: false,
        relatedWork: [
          {
            entityId: "story-1",
            reference: "WEB-12",
          },
        ],
      },
    });
    expect(result).not.toHaveProperty("document.contentHtml");
    expect(result).not.toHaveProperty("document.sharedWith");
  });

  it("bounds very large document content and marks it as incomplete", async () => {
    getDocumentMock.mockResolvedValue({
      ...summary("document-1"),
      contentHtml: "",
      contentText: "x".repeat(20_100),
      sharedWith: [],
      relatedWork: [],
    });

    const result = (await executeTool(getDocumentDetailsTool.execute, {
      documentId: "document-1",
    })) as {
      document: { content: string; contentTruncated: boolean };
    };

    expect(result.document.content).toHaveLength(20_000);
    expect(result.document.content.endsWith("...")).toBe(true);
    expect(result.document.contentTruncated).toBe(true);
  });

  it("does not query documents without authentication", async () => {
    authMock.mockResolvedValue(null);

    const result = await executeTool(listDocumentsTool.execute, {});

    expect(getDocumentsMock).not.toHaveBeenCalled();
    expect(result).toEqual({
      success: false,
      error: "Authentication required to access documents",
    });
  });
});
