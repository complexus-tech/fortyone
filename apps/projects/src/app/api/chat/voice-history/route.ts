import type { NextRequest } from "next/server";
import { ApiError } from "api-client";
import { z } from "zod";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getAiChatMessages } from "@/modules/ai-chats/queries/get-ai-chat-messages";
import { beginChatWrite, saveChat } from "../save-chat";
import { ChatAdmissionError } from "../request-context-policy";
import { appendVoiceHistory } from "./append";

const requestSchema = z
  .object({
    id: z.string().length(16),
    workspace: z
      .object({ slug: z.string().regex(/^[a-z0-9][a-z0-9-]{0,254}$/) })
      .strict(),
    messages: z
      .array(
        z
          .object({
            id: z.string().min(1).max(128),
            role: z.enum(["user", "assistant"]),
            text: z.string().trim().min(1).max(16_000),
          })
          .strict(),
      )
      .min(1)
      .max(40),
  })
  .strict();
const MAX_REQUEST_BYTES = 128 * 1024;

export async function POST(request: NextRequest) {
  try {
    const session = await auth();
    if (!session?.user) return new Response("Unauthorized", { status: 401 });
    const text = await request.text();
    if (new TextEncoder().encode(text).byteLength > MAX_REQUEST_BYTES)
      return new Response("Voice history is too large.", { status: 413 });
    let raw: unknown;
    try {
      raw = JSON.parse(text) as unknown;
    } catch {
      return new Response("Invalid voice history.", { status: 400 });
    }
    const parsed = requestSchema.safeParse(raw);
    if (!parsed.success)
      return new Response("Invalid voice history.", { status: 400 });
    const { id, workspace, messages } = parsed.data;
    const ctx = { session, workspaceSlug: workspace.slug };
    const authorizedWorkspace = await getWorkspace(ctx);
    if (!authorizedWorkspace.id || authorizedWorkspace.deletedAt)
      return new Response("Workspace access is unavailable.", { status: 403 });
    const saved = await appendVoiceHistory(messages, {
      read: async () => {
        try {
          return await getAiChatMessages(ctx, id);
        } catch (error) {
          if (error instanceof ApiError && error.status === 404) return [];
          throw error;
        }
      },
      begin: (history) =>
        beginChatWrite({
          id,
          messages: history,
          operation: "append",
          workspaceSlug: workspace.slug,
        }),
      finalize: (history, reservation) =>
        saveChat({
          id,
          messages: history,
          reservation,
          workspaceSlug: workspace.slug,
        }),
    });
    return Response.json(
      { messages: saved },
      { headers: { "Cache-Control": "no-store" } },
    );
  } catch (error) {
    if (error instanceof ChatAdmissionError)
      return new Response(error.message, { status: error.status });
    if (
      error instanceof ApiError &&
      [401, 403, 404, 409].includes(error.status)
    )
      return new Response(
        "Voice history could not be saved. Reload the conversation and retry.",
        { status: error.status },
      );
    return new Response(
      "Voice history could not be saved. Retry the same transcripts.",
      { status: 503 },
    );
  }
}
