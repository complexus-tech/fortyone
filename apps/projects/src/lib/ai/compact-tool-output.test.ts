/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { compactToolOutput } from "./compact-tool-output";

describe("compactToolOutput", () => {
  it("removes UI-only fields and bounds arrays for model context", () => {
    const output = compactToolOutput({
      description: "Useful text",
      descriptionHTML: "<p>Useful text</p>",
      imageUrl: "https://example.com/image.png",
      items: Array.from({ length: 30 }, (_, index) => ({ id: index })),
      success: true,
    }) as Record<string, unknown>;

    expect(output.description).toBe("Useful text");
    expect(output).not.toHaveProperty("descriptionHTML");
    expect(output).not.toHaveProperty("imageUrl");
    expect(output.items).toHaveLength(20);
    expect(output.modelItemsOmitted).toEqual({ items: 10 });
  });

  it("truncates long strings and removes encoded binary payloads", () => {
    const output = compactToolOutput({
      attachment: `data:image/png;base64,${"a".repeat(2000)}`,
      content: "x".repeat(2000),
    }) as Record<string, string>;

    expect(output.attachment).toBe("[binary data omitted]");
    expect(output.content.length).toBeLessThan(1300);
    expect(output.content.endsWith("…")).toBe(true);
  });

  it("preserves all 50 compact bulk-create receipts without rich records", () => {
    const output = compactToolOutput(
      {
        createdCount: 50,
        stories: Array.from({ length: 50 }, (_, index) => ({
          description: "Do not send this rich field to the model.",
          id: `story-${index + 1}`,
          title: `Test ${index + 1}`,
        })),
        success: true,
      },
      { toolName: "bulkCreateStories" },
    ) as Record<string, unknown>;

    expect(output.stories).toHaveLength(50);
    expect((output.stories as unknown[])[49]).toEqual({
      id: "story-50",
      title: "Test 50",
    });
    expect(output.modelItemsOmitted).toBeUndefined();
  });

  it("preserves all 50 bulk-delete IDs when recompacting chat history", () => {
    const storyIds = Array.from(
      { length: 50 },
      (_, index) => `story-${index + 1}`,
    );
    const output = compactToolOutput({
      deletedCount: 50,
      requestedCount: 50,
      storyIds,
      success: true,
    }) as Record<string, unknown>;

    expect(output.storyIds).toEqual(storyIds);
    expect(output.modelItemsOmitted).toBeUndefined();
  });

  it("preserves all missing bulk-delete IDs for an exact retry receipt", () => {
    const missingStoryIds = Array.from(
      { length: 50 },
      (_, index) => `missing-story-${index + 1}`,
    );
    const output = compactToolOutput(
      {
        deletedCount: 0,
        missingStoryIds,
        requestedCount: 50,
        storyIds: [],
        success: false,
      },
      { toolName: "bulkDeleteStories" },
    ) as Record<string, unknown>;

    expect(output.missingStoryIds).toEqual(missingStoryIds);
    expect(output.modelItemsOmitted).toBeUndefined();
  });

  it("preserves bounded document content without retaining rich UI fields", () => {
    const content = "x".repeat(20_000);
    const output = compactToolOutput(
      {
        success: true,
        document: {
          id: "document-1",
          title: "Launch brief",
          content,
          contentHTML: `<p>${content}</p>`,
          contentTruncated: false,
          imageUrl: "https://storage.example.com/private-preview.png",
          relatedWork: Array.from({ length: 50 }, (_, index) => ({
            entityId: `story-${index + 1}`,
            entityType: "story",
            title: `Story ${index + 1}`,
          })),
        },
      },
      { toolName: "getDocumentDetailsTool" },
    ) as {
      document: Record<string, unknown>;
    };

    expect(output.document.content).toBe(content);
    expect(output.document.contentHTML).toBeUndefined();
    expect(output.document.imageUrl).toBeUndefined();
    expect(JSON.stringify(output).length).toBeLessThanOrEqual(24_000);
  });

  it("preserves bounded document content when recompacting chat history", () => {
    const content = "Document context ".repeat(1000).slice(0, 16_000);
    const output = compactToolOutput({
      success: true,
      document: {
        id: "document-1",
        title: "Launch brief",
        content,
        contentTruncated: false,
      },
    }) as { document: Record<string, unknown> };

    expect(output.document.content).toBe(content);
  });

  it("redacts Google Drive content from historical model context", () => {
    const content = "Drive context ".repeat(1600).slice(0, 20_000);
    const output = compactToolOutput(
      {
        success: true,
        file: {
          referenceId: "00000000-0000-4000-8000-000000000001",
          name: "Launch plan",
          mimeType: "application/vnd.google-apps.document",
          webViewLink: "https://docs.google.com/document/d/example/edit",
          content,
          contentType: "text/plain",
          contentTruncated: false,
          bytesRead: 20_000,
          untrustedExternalContent: true,
        },
      },
      { toolName: "getLinkedGoogleFileContentTool" },
    ) as { file: Record<string, unknown> };

    expect(output.file.content).toBeUndefined();
    expect(output.file.contentRetained).toBe(false);
    expect(output.file.untrustedExternalContent).toBe(true);
    expect(JSON.stringify(output)).not.toContain(content);
    expect(JSON.stringify(output).length).toBeLessThanOrEqual(24_000);
  });

  it("redacts attachment storage URLs and member emails only", () => {
    const output = compactToolOutput(
      {
        success: true,
        attachments: [
          {
            id: "attachment-1",
            filename: "brief.pdf",
            mimeType: "application/pdf",
            size: 1024,
            url: "https://storage.example.com/private-signed-url",
            uploadedBy: {
              id: "user-1",
              email: "joseph@example.com",
              fullName: "Joseph Mukorivo",
              username: "joseph",
              profileUrl: "/complexus/members/joseph",
            },
          },
        ],
      },
      { toolName: "listAttachments" },
    ) as {
      attachments: Record<string, unknown>[];
    };

    expect(output.attachments[0]).not.toHaveProperty("url");
    expect(output.attachments[0]?.uploadedBy).toEqual({
      id: "user-1",
      fullName: "Joseph Mukorivo",
      username: "joseph",
      profileUrl: "/complexus/members/joseph",
    });
  });

  it("keeps GitHub install credentials out of model output without mutating UI output", () => {
    const rawOutput = {
      success: true,
      installUrl:
        "https://github.com/apps/fortyone/installations/new?state=signed-session-value&token=secret-token",
      token: "secret-token",
      signedSession: {
        value: "signed-session-value",
      },
      message: "GitHub install session created.",
    };

    const output = compactToolOutput(rawOutput, {
      toolName: "createGitHubInstallSessionTool",
    });

    expect(output).toEqual({
      success: true,
      installSessionReady: true,
      message:
        "GitHub install session created. Continue using the link shown in the interface.",
    });
    expect(JSON.stringify(output)).not.toContain("secret-token");
    expect(JSON.stringify(output)).not.toContain("signed-session-value");
    expect(rawOutput.installUrl).toContain("signed-session-value");
  });

  it("retains analytics aggregates and bounded ranked data above the context budget", () => {
    const report = {
      sectionErrors: {
        engagement: "Engagement data is temporarily unavailable.",
      },
      workload: {
        summary: {
          totalOpenStories: 180,
          overdueStories: 12,
          unestimatedStories: 38,
        },
        members: Array.from({ length: 500 }, (_, index) => ({
          userId: `user-${index + 1}`,
          username: `person-${index + 1}`,
          openStories: 500 - index,
          context: "x".repeat(500),
        })),
      },
      requests: {
        totalRequests: 72,
        sources: Array.from({ length: 500 }, (_, index) => ({
          source: `source-${index + 1}`,
          count: index + 1,
          context: "y".repeat(500),
        })),
      },
    };

    const output = compactToolOutput(
      {
        success: true,
        kind: "workspace-command-center-report",
        title: "Workspace analytics command center",
        report,
      },
      { toolName: "workspaceCommandCenterReportTool" },
    ) as Record<string, unknown>;

    expect(output).toHaveProperty("report");
    expect(output).toHaveProperty(
      "report.workload.summary.totalOpenStories",
      180,
    );
    expect(output).toHaveProperty(
      "report.sectionErrors.engagement",
      "Engagement data is temporarily unavailable.",
    );
    expect(output).toHaveProperty("report.workload.modelItemsOmitted.members");
    expect(JSON.stringify(output).length).toBeLessThanOrEqual(24_000);
  });
});
