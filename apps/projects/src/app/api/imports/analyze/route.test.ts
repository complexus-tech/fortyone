/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
/* eslint-disable turbo/no-undeclared-env-vars -- the route reads its server-only OpenAI configuration */

import { createHash } from "node:crypto";
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
        sourceNamespace: null,
        teams: [],
        people: [],
        labels: [],
        strategicPillars: [],
        objectives: [],
        keyResults: [],
        sprints: [],
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
          source_type: "jira_csv",
          workspace_id: "workspace-1",
        }),
        max_output_tokens: 64_000,
        model: "gpt-5.6-terra",
        store: true,
      }),
    );
    expect(
      JSON.stringify(mockResponsesCreate.mock.calls[0]?.[0]?.input),
    ).toContain("the first data record is row-2");
    expect(mockResponsesCreate.mock.calls[0]?.[0]?.metadata).not.toHaveProperty(
      "source_namespace",
    );
  });

  it("returns a deterministic JSON preview when AI is not configured", async () => {
    delete process.env.OPENAI_API_KEY;

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          id: "board-1",
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
          prefs: {},
        }),
        fileName: "trello.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      analysis: {
        sourceType: "json",
        sourceNamespace: "trello:board:board-1",
        tasks: [
          {
            sourceId: "65a1234567890abcdef12345",
            status: "Doing",
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

  it("sends a compact normalized Trello graph instead of the raw export", async () => {
    mockResponsesCreate.mockResolvedValue({ id: "resp_trello_import_1" });

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          actions: [
            {
              data: {
                text: "raw-action-history-must-not-be-sent",
              },
            },
          ],
          cards: [
            {
              closed: false,
              desc: "Move this card",
              id: "card-1",
              idList: "list-1",
              name: "Imported Trello card",
            },
          ],
          id: "board-1",
          lists: [{ closed: false, id: "list-1", name: "Doing" }],
          name: "Migration board",
          prefs: {},
        }),
        fileName: "trello-export.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      responseId: "resp_trello_import_1",
      status: "queued",
    });
    const request = mockResponsesCreate.mock.calls[0]?.[0];
    expect(request).toEqual(
      expect.objectContaining({ model: "gpt-5.6-terra" }),
    );
    const serializedInput = JSON.stringify(request?.input);
    expect(serializedInput).toContain(
      '"filename":"trello-export.normalized.json"',
    );
    expect(serializedInput).not.toContain(
      "raw-action-history-must-not-be-sent",
    );
    expect(serializedInput).toContain(
      "return a task record only when you can add a credible semantic enrichment",
    );
    expect(serializedInput).toContain("Do not echo unchanged tasks");

    const content = request?.input?.[0]?.content as
      | { file_data?: string; type: string }[]
      | undefined;
    const fileData = content?.find(
      (item) => item.type === "input_file",
    )?.file_data;
    expect(fileData).toMatch(/^data:application\/json;base64,/u);
    const normalizedGraph = JSON.parse(
      Buffer.from(fileData?.split(",")[1] ?? "", "base64").toString("utf8"),
    ) as Record<string, unknown>;
    expect(normalizedGraph).toMatchObject({
      authoritativeTaskGraph: true,
      format: "fortyone-normalized-import-graph-v1",
      sourceKind: "trello",
      tasks: [{ sourceId: "card-1", title: "Imported Trello card" }],
    });
    expect(normalizedGraph).not.toHaveProperty("actions");
  });

  it("reports a context-limit queue failure while preserving the deterministic preview", async () => {
    mockResponsesCreate.mockRejectedValue(
      Object.assign(new Error("maximum context window"), {
        code: "context_length_exceeded",
        status: 400,
      }),
    );

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          cards: [{ id: "card-1", idList: "list-1", name: "Keep me" }],
          id: "board-1",
          lists: [{ id: "list-1", name: "Doing" }],
          prefs: {},
        }),
        fileName: "trello.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      analysis: {
        tasks: [{ sourceId: "card-1", title: "Keep me" }],
        warnings: expect.arrayContaining([
          "The source exceeded the AI analysis context limit. The deterministic import preview is still available.",
        ]),
      },
      responseId: null,
      status: "completed",
    });
  });

  it("queues an entity-only JSON document for vendor-neutral AI mapping", async () => {
    mockResponsesCreate.mockResolvedValue({ id: "resp_json_import_1" });

    const response = await POST(
      createPostRequest({
        contents: JSON.stringify({
          containerId: "portfolio-1",
          projects: [{ id: "project-1", name: "Improve activation" }],
          teams: [{ id: "team-1", name: "Growth" }],
          members: [{ email: "owner@example.com", id: "member-1" }],
        }),
        fileName: "portfolio.json",
        mimeType: "application/json",
      }),
    );

    await expect(response.json()).resolves.toMatchObject({
      analysis: { sourceNamespace: null, sourceType: "json" },
      responseId: "resp_json_import_1",
      status: "queued",
    });
    const request = mockResponsesCreate.mock.calls[0]?.[0];
    expect(JSON.stringify(request?.input)).toContain(
      '"filename":"portfolio.json"',
    );
    expect(request?.metadata).toEqual(
      expect.objectContaining({
        source_type: "json",
      }),
    );
    expect(request?.metadata).not.toHaveProperty("source_namespace");
    expect(JSON.stringify(request?.input)).toContain(
      "The source format and product are unknown",
    );
    expect(JSON.stringify(request?.input)).not.toContain(
      "Do not echo unchanged tasks",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "source projects, initiatives, epics, or goals as objectives",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "source workspaces or teams may be teams",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "source portfolios, programs, strategic themes, or equivalent groupings as strategic pillars",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "set pillarSourceId to that pillar's sourceId",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "include only explicit measurable outcomes",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "include only true timeboxes",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Never derive or invent an email from a name",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Durable source namespaces are assigned only from trusted server-side parser metadata",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "sourceNamespace: always return null",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "first person in stable source order as the primary assignee",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Never repeat the primary assignee in collaborators",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "explicit shortSummary (at most 500 characters)",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "preserve explicit checklists, acceptance criteria, useful custom fields",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Never fetch attachment content",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "canonical source card, issue, or task URL",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "explicit remote attachment URLs as task links",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Never fetch attachment content or download any linked content",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "relative, data, file, or javascript URLs",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "source comments, activity, attachment bodies, estimates",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "map only explicit source issue or card relationships",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "Never infer a reciprocal relationship, invent an association, or create a self-link",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "keep checklist items inside their parent task description by default",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "A stable checklist-item ID alone does not make it a separate task",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "only preserve effort values explicitly represented by the source",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "never infer a unit for a bare or ambiguous number",
    );
    expect(JSON.stringify(request?.input)).toContain(
      "minimumFocusBlockMinutes requires estimatedDurationMinutes and cannot exceed it",
    );
  });

  it("normalizes every entity collection and source relationship", async () => {
    const fileHash = "a".repeat(64);
    const actorHash = createHash("sha256")
      .update(session.user.id)
      .digest("hex")
      .slice(0, 48);
    mockResponsesRetrieve.mockResolvedValue({
      metadata: {
        actor_hash: actorHash,
        file_hash: fileHash,
        fortyone_kind: "work_import_analysis",
        source_namespace: "trello:board:board-1",
        source_type: "json",
        workspace_id: "workspace-1",
      },
      output_text: JSON.stringify({
        sourceType: "jira_csv",
        sourceNamespace: " ai:container:wrong ",
        summary: " Structured work graph ",
        warnings: [" Review an ambiguous owner. "],
        mapping: null,
        teams: [
          {
            sourceId: " team-1 ",
            name: " Product ",
            code: " PROD ",
            color: " #3366FF ",
            description: " Delivery team ",
            isPrivate: false,
          },
        ],
        people: [
          {
            sourceId: " person-1 ",
            name: " Owner ",
            email: " OWNER@EXAMPLE.COM ",
            teamSourceIds: [" team-1 ", "team-1"],
          },
          {
            sourceId: " person-2 ",
            name: " Collaborator ",
            email: null,
            teamSourceIds: [" team-1 "],
          },
        ],
        labels: [
          {
            sourceId: " label-1 ",
            name: " Migration ",
            color: " blue ",
            teamSourceId: " team-1 ",
          },
          {
            sourceId: " duplicate-label ",
            name: " Duplicate one ",
            color: null,
            teamSourceId: null,
          },
          {
            sourceId: "duplicate-label",
            name: "Duplicate two",
            color: null,
            teamSourceId: null,
          },
        ],
        strategicPillars: [
          {
            sourceId: " pillar-1 ",
            name: " Customer growth ",
            description: " Grow active customers ",
            orderIndex: 0,
          },
        ],
        objectives: [
          {
            sourceId: " objective-1 ",
            name: " Complete migration ",
            description: " Move active work ",
            shortSummary: " Move safely ",
            color: " #112233 ",
            isPrivate: false,
            status: " In progress ",
            statusCategory: "started",
            priority: "High",
            leadPersonSourceId: " person-1 ",
            teamSourceId: " team-1 ",
            pillarSourceId: " pillar-1 ",
            startDate: "2026-02-31",
            endDate: "2026-09-30",
          },
          {
            sourceId: " objective-2 ",
            name: " Improve reliability ",
            description: null,
            shortSummary: null,
            color: null,
            isPrivate: true,
            status: null,
            statusCategory: null,
            priority: "No Priority",
            leadPersonSourceId: null,
            teamSourceId: " team-1 ",
            pillarSourceId: null,
            startDate: null,
            endDate: null,
          },
        ],
        keyResults: [
          {
            sourceId: " kr-1 ",
            name: " Move every card ",
            objectiveSourceId: " objective-2 ",
            measurementType: "percentage",
            startValue: 0,
            currentValue: 25,
            targetValue: 100,
            leadPersonSourceId: " person-1 ",
            contributorPersonSourceIds: [" person-1 ", "person-1"],
            startDate: "2026-09-01",
            endDate: "2026-09-30",
          },
        ],
        sprints: [
          {
            sourceId: " sprint-1 ",
            name: " Migration week ",
            goal: " Move active work ",
            teamSourceId: " team-1 ",
            objectiveSourceId: " objective-1 ",
            startDate: "2026-09-01",
            endDate: "2026-13-01",
          },
        ],
        tasks: [
          {
            sourceId: " story-1 ",
            title: " Move the backlog ",
            description: " Preserve this task. ",
            status: " Doing ",
            statusCategory: "started",
            priority: "High",
            estimateValue: 5,
            estimatedDurationMinutes: 90,
            minimumFocusBlockMinutes: 30,
            assigneeEmail: " OWNER@EXAMPLE.COM ",
            assigneeName: " Owner ",
            assigneePersonSourceId: " person-1 ",
            collaboratorPersonSourceIds: [
              " person-1 ",
              " person-2 ",
              "person-2",
            ],
            teamSourceId: " team-1 ",
            parentSourceId: " missing-parent ",
            objectiveSourceId: " objective-1 ",
            keyResultSourceId: " kr-1 ",
            sprintSourceId: " sprint-1 ",
            labelSourceIds: [" label-1 ", "label-1"],
            associations: [
              { type: "related", targetSourceId: " story-2 " },
              { type: "related", targetSourceId: "story-2" },
              { type: "blocks", targetSourceId: " story-1 " },
              { type: "blocked_by", targetSourceId: " missing-story " },
            ],
            links: [
              {
                title: " Original Trello card ",
                url: "HTTPS://EXAMPLE.COM:443/card/1",
              },
              {
                title: "Duplicate card",
                url: "https://example.com/card/1",
              },
              {
                title: " Attachment ",
                url: "https://cdn.example.com/export.pdf",
              },
              {
                title: "Unsafe script",
                url: ["javascript", "alert(1)"].join(":"),
              },
              { title: "Unsafe data", url: "data:text/plain,unsafe" },
              { title: "Relative", url: "/cards/1" },
            ],
            startDate: "2026-09-01",
            endDate: "2026-02-31",
          },
          {
            sourceId: " story-2 ",
            title: " Verify the migration ",
            description: " Confirm imported work. ",
            status: null,
            statusCategory: "unstarted",
            priority: "Medium",
            estimateValue: "medium",
            estimatedDurationMinutes: 60,
            minimumFocusBlockMinutes: 90,
            assigneeEmail: null,
            assigneeName: " Collaborator ",
            assigneePersonSourceId: " person-2 ",
            collaboratorPersonSourceIds: [],
            teamSourceId: " team-1 ",
            parentSourceId: null,
            objectiveSourceId: " objective-2 ",
            keyResultSourceId: " kr-1 ",
            sprintSourceId: null,
            labelSourceIds: [],
            associations: [],
            links: [],
            startDate: null,
            endDate: null,
          },
        ],
      }),
      status: "completed",
    });

    const response = await GET({
      url: `http://localhost/api/imports/analyze?workspaceSlug=acme&responseId=resp_import_1&fileHash=${fileHash}`,
    } as Request);

    const body = (await response.json()) as {
      analysis: { warnings: string[] };
    };

    expect(body).toMatchObject({
      analysis: {
        sourceType: "json",
        sourceNamespace: "trello:board:board-1",
        teams: [
          {
            sourceId: "team-1",
            name: "Product",
            code: "PROD",
            color: "#3366FF",
            description: "Delivery team",
            isPrivate: false,
          },
        ],
        people: [
          {
            email: "owner@example.com",
            teamSourceIds: ["team-1"],
          },
          {
            sourceId: "person-2",
            name: "Collaborator",
            teamSourceIds: ["team-1"],
          },
        ],
        labels: [{ teamSourceId: "team-1" }],
        strategicPillars: [
          {
            sourceId: "pillar-1",
            name: "Customer growth",
            description: "Grow active customers",
            orderIndex: 0,
          },
        ],
        objectives: [
          {
            shortSummary: "Move safely",
            color: "#112233",
            isPrivate: false,
            leadPersonSourceId: "person-1",
            pillarSourceId: "pillar-1",
            startDate: null,
            endDate: "2026-09-30",
          },
          {
            sourceId: "objective-2",
            isPrivate: true,
          },
        ],
        keyResults: [
          {
            objectiveSourceId: "objective-2",
            contributorPersonSourceIds: ["person-1"],
          },
        ],
        sprints: [{ objectiveSourceId: "objective-1", endDate: null }],
        tasks: [
          {
            assigneeEmail: "owner@example.com",
            assigneeName: "Owner",
            assigneePersonSourceId: "person-1",
            collaboratorPersonSourceIds: ["person-2"],
            estimateValue: 5,
            estimatedDurationMinutes: 90,
            minimumFocusBlockMinutes: 30,
            labelSourceIds: ["label-1"],
            associations: [{ type: "related", targetSourceId: "story-2" }],
            links: [
              {
                title: "Original Trello card",
                url: "https://example.com/card/1",
              },
              {
                title: "Attachment",
                url: "https://cdn.example.com/export.pdf",
              },
            ],
            endDate: null,
          },
          {
            sourceId: "story-2",
            assigneePersonSourceId: "person-2",
            associations: [],
            links: [],
            estimateValue: null,
            estimatedDurationMinutes: 60,
            minimumFocusBlockMinutes: null,
          },
        ],
      },
      status: "completed",
    });
    expect(body.analysis.warnings).toEqual(
      expect.arrayContaining([
        "2 source objects were omitted because 1 source ID was duplicated and could not be related safely.",
        "1 duplicate task association was deduplicated.",
        "1 self-referential task association was removed.",
        "1 task association targeting an unreturned task was removed.",
        "1 source relationship points to objects that were not returned and will use safe fallbacks.",
        "1 task relationship has conflicting objective and key-result references and needs review.",
        "1 source team description remains visible for review but cannot be applied by FortyOne's team creation contract.",
        "3 unsafe or malformed task links were omitted; only absolute HTTP or HTTPS URLs are supported.",
        "1 duplicate task link was deduplicated by canonical URL.",
        "1 invalid or ambiguous task complexity estimate was omitted; FortyOne accepts only explicit values 1, 2, 3, 5, 8.",
        "1 task minimum focus block was omitted because it exceeded the estimated duration.",
        "3 invalid calendar dates were omitted instead of being guessed.",
      ]),
    );

    const originalOpenAIResponse = (await mockResponsesRetrieve.mock.results[0]
      ?.value) as { output_text: string; [key: string]: unknown };
    const analysisWithoutAuthoritativeNamespace = JSON.parse(
      originalOpenAIResponse.output_text,
    ) as Record<string, unknown>;
    mockResponsesRetrieve.mockResolvedValue({
      ...originalOpenAIResponse,
      metadata: {
        actor_hash: actorHash,
        file_hash: fileHash,
        fortyone_kind: "work_import_analysis",
        source_type: "json",
        workspace_id: "workspace-1",
      },
      output_text: JSON.stringify({
        ...analysisWithoutAuthoritativeNamespace,
        sourceNamespace: " jira:site:site-1 ",
      }),
    });

    const responseWithoutAuthoritativeNamespace = await GET({
      url: `http://localhost/api/imports/analyze?workspaceSlug=acme&responseId=resp_import_2&fileHash=${fileHash}`,
    } as Request);
    await expect(
      responseWithoutAuthoritativeNamespace.json(),
    ).resolves.toMatchObject({
      analysis: { sourceNamespace: null },
      status: "completed",
    });
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

  it("surfaces an incomplete background analysis reason", async () => {
    const fileHash = "a".repeat(64);
    const actorHash = createHash("sha256")
      .update(session.user.id)
      .digest("hex")
      .slice(0, 48);
    mockResponsesRetrieve.mockResolvedValue({
      incomplete_details: { reason: "max_output_tokens" },
      metadata: {
        actor_hash: actorHash,
        file_hash: fileHash,
        fortyone_kind: "work_import_analysis",
        source_type: "json",
        workspace_id: "workspace-1",
      },
      status: "incomplete",
    });

    const response = await GET({
      url: `http://localhost/api/imports/analyze?workspaceSlug=acme&responseId=resp_import_1&fileHash=${fileHash}`,
    } as Request);

    expect(response.status).toBe(502);
    await expect(response.text()).resolves.toBe(
      "AI analysis reached its output limit before finishing. The deterministic import preview is still available.",
    );
  });
});
