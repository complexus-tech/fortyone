/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
/* eslint-disable turbo/no-undeclared-env-vars -- the route reads its server-only OpenAI configuration */

import OpenAI from "openai";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { GET, POST } from "./route";

const mockResponsesCreate = jest.fn();
const mockResponsesRetrieve = jest.fn();

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("openai", () => ({
  __esModule: true,
  default: jest.fn(() => ({
    responses: {
      create: mockResponsesCreate,
      retrieve: mockResponsesRetrieve,
    },
  })),
}));
jest.mock("openai/helpers/zod", () => ({
  zodTextFormat: jest.fn(() => ({ type: "json_schema" })),
}));

const mockedAuth = jest.mocked(auth);
const mockedGetWorkspace = jest.mocked(getWorkspace);
const mockedOpenAI = jest.mocked(OpenAI);

const session = {
  user: {
    email: "admin@example.com",
    fullName: "Example Admin",
    id: "user-1",
    image: null,
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
    name: "Example Admin",
    username: "admin",
  },
};

class MockResponse {
  readonly headers: Headers;
  readonly status: number;

  constructor(
    readonly body?: unknown,
    init?: { headers?: HeadersInit; status?: number },
  ) {
    this.headers = new Headers(init?.headers);
    this.status = init?.status ?? 200;
  }

  static json(
    body: unknown,
    init?: { headers?: HeadersInit; status?: number },
  ) {
    return new MockResponse(body, init);
  }

  async json() {
    return this.body;
  }

  async text() {
    return typeof this.body === "string"
      ? this.body
      : JSON.stringify(this.body);
  }
}

const createPostRequest = ({
  contents,
  fileName = "jira.csv",
  mimeType = "text/csv",
}: {
  contents: string;
  fileName?: string;
  mimeType?: string;
}) => {
  const formData = new FormData();
  const file = new File([contents], fileName, { type: mimeType });
  Object.defineProperty(file, "arrayBuffer", {
    configurable: true,
    value: async () => Uint8Array.from(Buffer.from(contents)).buffer,
  });
  formData.set("file", file);
  return {
    formData: jest.fn().mockResolvedValue(formData),
    headers: new Headers(),
    url: "http://localhost/api/imports/analyze?workspaceSlug=acme",
  } as unknown as Request;
};

describe("/api/imports/analyze", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    process.env.OPENAI_API_KEY = "test-key";
    (globalThis as { Response: typeof Response }).Response =
      MockResponse as never;
    mockedAuth.mockResolvedValue(session);
    mockedGetWorkspace.mockResolvedValue({
      id: "workspace-1",
      userRole: "admin",
    } as never);
  });

  it("authenticates before parsing an untrusted multipart body", async () => {
    mockedAuth.mockResolvedValue(null);
    const formData = jest.fn();

    const response = await POST({ formData } as unknown as Request);

    expect(response.status).toBe(401);
    expect(formData).not.toHaveBeenCalled();
    expect(mockedGetWorkspace).not.toHaveBeenCalled();
  });

  it("authorizes the workspace and rejects oversized requests before parsing multipart data", async () => {
    const formData = jest.fn();
    const response = await POST({
      formData,
      headers: new Headers({ "content-length": String(21 * 1024 * 1024) }),
      url: "http://localhost/api/imports/analyze?workspaceSlug=acme",
    } as unknown as Request);

    expect(response.status).toBe(413);
    expect(mockedGetWorkspace).toHaveBeenCalledWith({
      session,
      workspaceSlug: "acme",
    });
    expect(formData).not.toHaveBeenCalled();
  });

  it("parses a Jira CSV deterministically without sending it to OpenAI", async () => {
    delete process.env.OPENAI_API_KEY;

    const response = await POST(
      createPostRequest({
        contents:
          "Issue key,Summary,Status,Priority\nPROJ-42,Ship import,In Progress,High",
      }),
    );
    const body = (await response.json()) as {
      analysis: { sourceType: string; tasks: { sourceId: string }[] };
      status: string;
    };

    expect(response.status).toBe(200);
    expect(response.headers.get("Cache-Control")).toBe("private, no-store");
    expect(body).toMatchObject({
      analysis: {
        sourceType: "jira_csv",
        tasks: [{ sourceId: "PROJ-42" }],
      },
      status: "completed",
    });
    expect(mockedOpenAI).not.toHaveBeenCalled();
    expect(mockResponsesCreate).not.toHaveBeenCalled();
  });

  it("automatically queues OpenAI analysis for a CSV upload", async () => {
    mockResponsesCreate.mockResolvedValue({ id: "resp_import_1" });

    const response = await POST(
      createPostRequest({
        contents: "Issue key,Summary\nPROJ-42,Ship import",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      responseId: "resp_import_1",
      status: "queued",
    });
    expect(mockResponsesCreate).toHaveBeenCalledWith(
      expect.objectContaining({
        background: true,
        metadata: expect.objectContaining({
          fortyone_kind: "work_import_analysis",
          workspace_id: "workspace-1",
        }),
        store: true,
      }),
    );
  });

  it("returns a deterministic JSON preview when AI is not configured", async () => {
    delete process.env.OPENAI_API_KEY;

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          cards: [
            {
              desc: "Move this card",
              id: "65a1234567890abcdef12345",
              idList: "list-1",
              name: "Imported Trello card",
            },
          ],
          lists: [{ id: "list-1", name: "Doing" }],
          name: "Migration board",
        }),
        fileName: "trello.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      analysis: {
        sourceType: "json",
        tasks: [
          {
            sourceId: "65a1234567890abcdef12345",
            status: null,
            title: "Imported Trello card",
          },
        ],
      },
      responseId: null,
      status: "completed",
    });
    expect(mockedOpenAI).not.toHaveBeenCalled();
    expect(mockResponsesCreate).not.toHaveBeenCalled();
  });

  it("queues the complete JSON document for vendor-neutral AI mapping", async () => {
    mockResponsesCreate.mockResolvedValue({ id: "resp_json_import_1" });

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          cards: [
            {
              id: "card-1",
              idList: "list-1",
              idMembers: ["member-1"],
              name: "Map this task",
            },
          ],
          lists: [{ id: "list-1", name: "Doing" }],
          members: [{ email: "owner@example.com", id: "member-1" }],
        }),
        fileName: "board.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      analysis: { sourceType: "json" },
      responseId: "resp_json_import_1",
      status: "queued",
    });
    const request = mockResponsesCreate.mock.calls[0]?.[0];
    expect(JSON.stringify(request?.input)).toContain('"filename":"board.json"');
    expect(JSON.stringify(request?.input)).toContain(
      "The source format and product are unknown",
    );
  });

  it("does not reveal another workspace's background analysis", async () => {
    mockResponsesRetrieve.mockResolvedValue({
      metadata: {
        actor_hash: "different-actor",
        file_hash: "a".repeat(64),
        fortyone_kind: "work_import_analysis",
        workspace_id: "workspace-1",
      },
      status: "completed",
    });
    const request = {
      url: `http://localhost/api/imports/analyze?workspaceSlug=acme&responseId=resp_import_1&fileHash=${"a".repeat(64)}`,
    } as Request;

    const response = await GET(request);

    expect(response.status).toBe(404);
    expect(response.headers.get("Cache-Control")).toBe("private, no-store");
  });
});
