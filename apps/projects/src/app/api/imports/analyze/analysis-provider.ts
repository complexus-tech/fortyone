/* eslint-disable turbo/no-undeclared-env-vars -- OPENAI_API_KEY is server-only */

import "server-only";

import type { ResponseInputContent } from "openai/resources/responses/responses";
import OpenAI from "openai";
import { zodTextFormat } from "openai/helpers/zod";
import {
  OPENAI_DEFAULT_REASONING_EFFORT,
  OPENAI_IMPORT_ANALYSIS_MODEL,
} from "@/lib/ai/models";
import {
  importAnalysisSchema,
  importSourceTypeSchema,
  type ImportSourceType,
} from "@/modules/settings/workspace/imports/schema";
import { createAnalysisPrompt } from "./analysis-prompt";
import {
  delimitedExtensions,
  digest,
  imageExtensions,
} from "./analysis-source";
import {
  normalizeAnalysis,
  normalizeSourceNamespace,
} from "./analysis-normalization";
import {
  normalizeDecodedTaskEffort,
  normalizeDecodedTaskLinks,
} from "./analysis-task-normalization";
import { jsonResponse, textResponse } from "./responses";

const IMPORT_AI_MAX_OUTPUT_TOKENS = 64_000;

export const getAIAnalysisFailureMessage = (
  error: unknown,
  fallback: string,
): string => {
  const failure =
    error && typeof error === "object"
      ? (error as Record<string, unknown>)
      : null;
  const nestedError =
    failure?.error && typeof failure.error === "object"
      ? (failure.error as Record<string, unknown>)
      : null;
  const incompleteDetails =
    failure?.incomplete_details &&
    typeof failure.incomplete_details === "object"
      ? (failure.incomplete_details as Record<string, unknown>)
      : null;
  const status =
    typeof failure?.status === "number" ? failure.status : undefined;
  const reason = [
    failure?.code,
    failure?.message,
    nestedError?.code,
    nestedError?.message,
    incompleteDetails?.reason,
  ]
    .filter((value): value is string => typeof value === "string")
    .join(" ")
    .toLowerCase();

  if (
    reason.includes("context_length") ||
    reason.includes("context window") ||
    reason.includes("too many tokens")
  ) {
    return "The source exceeded the AI analysis context limit. The deterministic import preview is still available.";
  }
  if (
    status === 413 ||
    reason.includes("request_too_large") ||
    reason.includes("payload too large")
  ) {
    return "The source was too large to send for AI enrichment. The deterministic import preview is still available.";
  }
  if (
    incompleteDetails?.reason === "max_output_tokens" ||
    reason.includes("max_output_tokens") ||
    reason.includes("max_tokens")
  ) {
    return "AI analysis reached its output limit before finishing. The deterministic import preview is still available.";
  }
  if (status === 429 || reason.includes("rate_limit")) {
    return "AI analysis is temporarily busy. The deterministic import preview is still available.";
  }
  if (reason.includes("content_filter") || reason.includes("invalid_prompt")) {
    return "AI analysis could not process this source safely. The deterministic import preview is still available.";
  }

  return fallback;
};

export const createBackgroundAnalysis = async ({
  actorHash,
  authoritativeTaskGraph,
  bytes,
  extension,
  fileHash,
  fileName,
  mimeType,
  sourceNamespace,
  sourceType,
  workspaceId,
}: {
  actorHash: string;
  authoritativeTaskGraph: boolean;
  bytes: Buffer;
  extension: string;
  fileHash: string;
  fileName: string;
  mimeType: string;
  sourceNamespace: string | undefined;
  sourceType: ImportSourceType;
  workspaceId: string;
}) => {
  const client = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
  const dataUrl = `data:${mimeType};base64,${bytes.toString("base64")}`;
  const fileContent: ResponseInputContent = imageExtensions.has(extension)
    ? { type: "input_image", detail: "high", image_url: dataUrl }
    : {
        type: "input_file",
        file_data: dataUrl,
        filename: fileName,
      };

  return client.responses.create({
    background: true,
    input: [
      {
        role: "user",
        content: [
          {
            type: "input_text",
            text: createAnalysisPrompt({
              authoritativeTaskGraph,
              delimited: delimitedExtensions.has(extension),
              sourceType,
            }),
          },
          fileContent,
        ],
      },
    ],
    max_output_tokens: IMPORT_AI_MAX_OUTPUT_TOKENS,
    metadata: {
      actor_hash: actorHash,
      file_hash: fileHash,
      fortyone_kind: "work_import_analysis",
      ...(sourceNamespace ? { source_namespace: sourceNamespace } : {}),
      source_type: sourceType,
      workspace_id: workspaceId,
    },
    model: OPENAI_IMPORT_ANALYSIS_MODEL,
    reasoning: { effort: OPENAI_DEFAULT_REASONING_EFFORT },
    safety_identifier: actorHash,
    store: true,
    text: {
      format: zodTextFormat(importAnalysisSchema, "work_import_analysis"),
      verbosity: "low",
    },
  });
};

export async function pollBackgroundAnalysis({
  actorId,
  fileHash,
  responseId,
  workspaceId,
}: {
  actorId: string;
  fileHash: string;
  responseId: string;
  workspaceId: string;
}): Promise<Response> {
  try {
    const client = new OpenAI({ apiKey: process.env.OPENAI_API_KEY });
    const response = await client.responses.retrieve(responseId);
    const actorHash = digest(actorId).slice(0, 48);
    if (
      response.metadata?.fortyone_kind !== "work_import_analysis" ||
      response.metadata.workspace_id !== workspaceId ||
      response.metadata.actor_hash !== actorHash ||
      response.metadata.file_hash !== fileHash
    ) {
      return textResponse("Import analysis not found", 404);
    }
    const sourceType = importSourceTypeSchema.safeParse(
      response.metadata.source_type,
    );
    if (!sourceType.success) {
      return textResponse("Import analysis not found", 404);
    }
    const metadataSourceNamespaceValue = (
      response.metadata as Record<string, unknown>
    ).source_namespace;
    if (
      metadataSourceNamespaceValue !== undefined &&
      typeof metadataSourceNamespaceValue !== "string"
    ) {
      return textResponse("Import analysis not found", 404);
    }
    const metadataSourceNamespace = normalizeSourceNamespace(
      metadataSourceNamespaceValue ?? null,
    );

    if (response.status === "queued" || response.status === "in_progress") {
      return jsonResponse({ status: response.status });
    }
    if (response.status !== "completed") {
      return textResponse(
        getAIAnalysisFailureMessage(
          response,
          "The AI analysis did not complete. The deterministic import preview is still available.",
        ),
        502,
      );
    }

    let decoded: unknown;
    try {
      decoded = JSON.parse(response.output_text) as unknown;
    } catch {
      return textResponse("The AI analysis returned an invalid result", 502);
    }
    if (!decoded || typeof decoded !== "object" || Array.isArray(decoded)) {
      return textResponse("The AI analysis returned an invalid result", 502);
    }
    const effortNormalization = normalizeDecodedTaskEffort(
      decoded as Record<string, unknown>,
    );
    const linkNormalization = normalizeDecodedTaskLinks(
      effortNormalization.decoded,
    );
    const analysis = importAnalysisSchema.safeParse({
      ...linkNormalization.decoded,
      sourceNamespace: metadataSourceNamespace,
      sourceType: sourceType.data,
    });
    if (!analysis.success) {
      return textResponse("The AI analysis returned an invalid result", 502);
    }

    return jsonResponse({
      analysis: normalizeAnalysis({
        ...analysis.data,
        warnings: [
          ...effortNormalization.warnings,
          ...linkNormalization.warnings,
          ...analysis.data.warnings,
        ],
      }),
      status: "completed",
    });
  } catch (error) {
    return textResponse(
      getAIAnalysisFailureMessage(
        error,
        "Unable to retrieve the AI analysis. The deterministic import preview is still available.",
      ),
      502,
    );
  }
}
