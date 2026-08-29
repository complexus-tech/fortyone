/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- The AI SDK requires web streams.
/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  asSchema,
  convertToModelMessages,
  tool,
  type ToolSet,
  type UIMessage,
} from "ai";
import { z } from "zod";
import { withCompactModelOutputs } from "./model-tools";

const needsApproval = async (
  registeredTool: ToolSet[string],
  input: unknown,
) => {
  const approval = registeredTool.needsApproval;
  return typeof approval === "function"
    ? approval(input, {
        experimental_context: {},
        messages: [],
        toolCallId: "tool-call-1",
      })
    : approval;
};

describe("withCompactModelOutputs", () => {
  it("keeps raw execution output while compacting the model-facing result", async () => {
    const rawOutput = {
      description: "Visible to the UI",
      descriptionHTML: `<p>${"large".repeat(500)}</p>`,
      success: true,
    };
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object({}),
        execute: () => rawOutput,
      }),
    });

    expect(toolSet.example.execute?.({}, {} as never)).toEqual(rawOutput);
    expect(
      await toolSet.example.toModelOutput?.({
        input: {},
        output: rawOutput,
        toolCallId: "tool-call-1",
      }),
    ).toEqual({
      type: "json",
      value: {
        description: "Visible to the UI",
        success: true,
      },
    });
  });

  it("compacts historical tool results when UI messages become model history", async () => {
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object({}),
        execute: () => ({
          descriptionHTML: `<p>${"large".repeat(500)}</p>`,
          success: true,
        }),
      }),
    });
    const messages = [
      {
        id: "assistant-1",
        role: "assistant",
        parts: [
          {
            input: {},
            output: {
              descriptionHTML: `<p>${"large".repeat(500)}</p>`,
              success: true,
            },
            state: "output-available",
            toolCallId: "tool-call-1",
            type: "tool-example",
          },
        ],
      },
    ] as UIMessage[];

    const modelMessages = await convertToModelMessages(messages, {
      tools: toolSet,
    });

    expect(modelMessages).toEqual([
      {
        content: [
          expect.objectContaining({
            toolCallId: "tool-call-1",
            toolName: "example",
            type: "tool-call",
          }),
        ],
        role: "assistant",
      },
      {
        content: [
          expect.objectContaining({
            output: { type: "json", value: { success: true } },
            toolCallId: "tool-call-1",
            toolName: "example",
            type: "tool-result",
          }),
        ],
        role: "tool",
      },
    ]);
  });

  it("keeps a GitHub install URL available to the UI but not the model", async () => {
    const rawOutput = {
      success: true,
      installUrl:
        "https://github.com/apps/fortyone/installations/new?state=signed-session",
      token: "secret-token",
      message: "GitHub install session created.",
    };
    const toolSet = withCompactModelOutputs({
      createGitHubInstallSessionTool: tool({
        inputSchema: z.object({}),
        execute: () => rawOutput,
      }),
    });

    expect(
      toolSet.createGitHubInstallSessionTool.execute?.({}, {} as never),
    ).toEqual(rawOutput);
    expect(
      await toolSet.createGitHubInstallSessionTool.toModelOutput?.({
        input: {},
        output: rawOutput,
        toolCallId: "tool-call-1",
      }),
    ).toEqual({
      type: "json",
      value: {
        success: true,
        installSessionReady: true,
        message:
          "GitHub install session created. Continue using the link shown in the interface.",
      },
    });

    const modelMessages = await convertToModelMessages(
      [
        {
          id: "assistant-github",
          role: "assistant",
          parts: [
            {
              input: {},
              output: rawOutput,
              state: "output-available",
              toolCallId: "github-install-call",
              type: "tool-createGitHubInstallSessionTool",
            },
          ],
        },
      ] as UIMessage[],
      { tools: toolSet },
    );
    const serializedHistory = JSON.stringify(modelMessages);

    expect(serializedHistory).toContain('"installSessionReady":true');
    expect(serializedHistory).not.toContain("signed-session");
    expect(serializedHistory).not.toContain("secret-token");
  });

  it("requires approval for named and dynamic mutations only", async () => {
    const toolSet = withCompactModelOutputs({
      comments: tool({
        inputSchema: z.object({ action: z.string() }),
        execute: () => ({ success: true }),
      }),
      createGitHubInstallSessionTool: tool({
        inputSchema: z.object({}),
        execute: () => ({ success: true }),
      }),
      listTeams: tool({
        inputSchema: z.object({}),
        execute: () => ({ success: true }),
      }),
      mayaWorkPlanTool: tool({
        inputSchema: z.object({ storyId: z.string() }),
        execute: () => ({ success: true }),
      }),
      applyMayaWorkPlanTool: tool({
        inputSchema: z.object({ runId: z.string() }),
        execute: () => ({ success: true }),
      }),
      updateStory: tool({
        inputSchema: z.object({ storyId: z.string() }),
        execute: () => ({ success: true }),
      }),
    });

    await expect(
      needsApproval(toolSet.updateStory, { storyId: "story-1" }),
    ).resolves.toBe(true);
    await expect(
      needsApproval(toolSet.comments, { action: "add-comment" }),
    ).resolves.toBe(true);
    await expect(
      needsApproval(toolSet.comments, { action: "list-comments" }),
    ).resolves.toBe(false);
    await expect(needsApproval(toolSet.listTeams, {})).resolves.toBe(false);
    await expect(
      needsApproval(toolSet.mayaWorkPlanTool, { storyId: "story-1" }),
    ).resolves.toBe(false);
    await expect(
      needsApproval(toolSet.applyMayaWorkPlanTool, { runId: "run-1" }),
    ).resolves.toBe(true);
    await expect(
      needsApproval(toolSet.createGitHubInstallSessionTool, {}),
    ).resolves.toBe(false);
  });

  it("removes only strict-provider null placeholders and preserves clear operations", async () => {
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object({
          optionalText: z.string().optional(),
          clearableText: z.string().nullable().optional(),
          items: z.array(
            z.object({
              optionalLabel: z.string().optional(),
              clearableLabel: z.string().nullable().optional(),
            }),
          ),
        }),
        execute: () => ({ success: true }),
      }),
    });
    const schema = asSchema(toolSet.example.inputSchema);

    await expect(
      schema.validate?.({
        optionalText: null,
        clearableText: null,
        items: [
          {
            optionalLabel: null,
            clearableLabel: null,
          },
        ],
      }),
    ).resolves.toEqual({
      success: true,
      value: {
        clearableText: null,
        items: [{ clearableLabel: null }],
      },
    });
  });

  it("still rejects a required field that the provider supplies as null", async () => {
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object({ requiredText: z.string() }),
        execute: () => ({ success: true }),
      }),
    });
    const schema = asSchema(toolSet.example.inputSchema);
    const result = await schema.validate?.({ requiredText: null });

    expect(result?.success).toBe(false);
  });

  it("normalizes more than eight optional null placeholders in one pass", async () => {
    const optionalFields = Object.fromEntries(
      Array.from({ length: 12 }, (_, index) => [
        `optionalField${index}`,
        z.string().optional(),
      ]),
    );
    const toolSet = withCompactModelOutputs({
      example: tool({
        inputSchema: z.object(optionalFields),
        execute: () => ({ success: true }),
      }),
    });
    const schema = asSchema(toolSet.example.inputSchema);
    const providerInput = Object.fromEntries(
      Object.keys(optionalFields).map((key) => [key, null]),
    );

    await expect(schema.validate?.(providerInput)).resolves.toEqual({
      success: true,
      value: {},
    });
  });
});
