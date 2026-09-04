/* eslint-disable turbo/no-undeclared-env-vars -- OPENAI_API_KEY is server-only */

import { z } from "zod";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { IMPORT_MAX_FILE_BYTES } from "@/modules/settings/workspace/imports/schema";
import { createDelimitedImportDraft } from "@/modules/settings/workspace/imports/csv";
import { createJsonImportDraft } from "@/modules/settings/workspace/imports/json";
import {
  acceptedExtensions,
  cleanFileName,
  createAIAnalysisFile,
  delimitedExtensions,
  digest,
  getFileExtension,
  getSourceType,
  jsonExtensions,
} from "./analysis-source";
import {
  createBackgroundAnalysis,
  getAIAnalysisFailureMessage,
  pollBackgroundAnalysis,
} from "./analysis-provider";
import { jsonResponse, textResponse } from "./responses";

export const maxDuration = 60;
export const runtime = "nodejs";

const maximumMultipartBytes = IMPORT_MAX_FILE_BYTES + 512 * 1024;

const workspaceSlugSchema = z
  .string()
  .trim()
  .min(1)
  .max(96)
  .regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/);

const pollQuerySchema = z.strictObject({
  responseId: z.string().regex(/^resp_[A-Za-z0-9_-]+$/),
  workspaceSlug: workspaceSlugSchema,
  fileHash: z.string().regex(/^[a-f0-9]{64}$/),
});

const getWorkspaceContext = async (
  workspaceSlug: string,
  session: NonNullable<Awaited<ReturnType<typeof auth>>>,
) => {
  try {
    const workspace = await getWorkspace({ session, workspaceSlug });
    if (workspace.userRole !== "admin") {
      return { error: textResponse("Forbidden", 403), ok: false } as const;
    }
    return { ok: true, session, workspace } as const;
  } catch {
    return {
      error: textResponse("Workspace not found", 404),
      ok: false,
    } as const;
  }
};

export async function POST(request: Request): Promise<Response> {
  const session = await auth();
  if (!session?.user) return textResponse("Unauthorized", 401);

  const workspaceSlug = workspaceSlugSchema.safeParse(
    new URL(request.url).searchParams.get("workspaceSlug"),
  );
  if (!workspaceSlug.success) {
    return textResponse("A valid workspace is required", 400);
  }
  const context = await getWorkspaceContext(workspaceSlug.data, session);
  if (!context.ok) return context.error;

  const contentLength = request.headers.get("content-length");
  if (contentLength) {
    const parsedLength = Number(contentLength);
    if (!Number.isSafeInteger(parsedLength) || parsedLength < 0) {
      return textResponse("Invalid import request length", 400);
    }
    if (parsedLength > maximumMultipartBytes) {
      return textResponse("The import request is too large", 413);
    }
  }

  let formData: FormData;
  try {
    formData = await request.formData();
  } catch {
    return textResponse("Invalid multipart import request", 400);
  }

  const entries = [...formData.entries()];
  const files = formData.getAll("file");
  const uploadedFiles = [...formData.values()].filter(
    (value): value is File => value instanceof File,
  );
  const file = files.at(0);
  if (
    entries.length !== 1 ||
    entries[0]?.[0] !== "file" ||
    files.length !== 1 ||
    uploadedFiles.length !== 1 ||
    !(file instanceof File)
  ) {
    return textResponse("Exactly one import file is required", 400);
  }

  const extension = getFileExtension(file.name);
  if (!acceptedExtensions.has(extension)) {
    return textResponse(
      "Use a CSV, TSV, JSON, Excel workbook, PDF, JPG, PNG, or WebP file",
      415,
    );
  }
  if (file.size <= 0 || file.size > IMPORT_MAX_FILE_BYTES) {
    return textResponse("The import file must be 20 MB or smaller", 413);
  }

  const bytes = Buffer.from(await file.arrayBuffer());
  const fileHash = digest(bytes);
  const fileName = cleanFileName(file.name);
  const sourceType = getSourceType(extension);
  let draft = null;

  if (delimitedExtensions.has(extension)) {
    try {
      draft = createDelimitedImportDraft({
        fileHash,
        fileName,
        text: bytes.toString("utf8"),
      });
    } catch (error) {
      return textResponse(
        error instanceof Error ? error.message : "Unable to read this file",
        400,
      );
    }
  }

  if (jsonExtensions.has(extension)) {
    try {
      draft = createJsonImportDraft({
        fileHash,
        fileName,
        text: bytes.toString("utf8"),
      });
    } catch (error) {
      return textResponse(
        error instanceof Error
          ? error.message
          : "Unable to read this JSON file",
        400,
      );
    }
  }
  const authoritativeSourceType = draft?.sourceType ?? sourceType;

  if (!process.env.OPENAI_API_KEY) {
    if (!draft) {
      return textResponse("AI file analysis is not configured", 503);
    }
    return jsonResponse({
      analysis: {
        ...draft,
        warnings: [
          ...draft.warnings,
          "AI mapping suggestions are unavailable, so the deterministic mapping is shown for review.",
        ],
      },
      fileHash,
      responseId: null,
      status: "completed",
    });
  }

  try {
    const actorHash = digest(context.session.user.id).slice(0, 48);
    const analysisFile = createAIAnalysisFile({
      bytes,
      draft,
      extension,
      fileName,
      mimeType: file.type || "application/octet-stream",
    });
    const response = await createBackgroundAnalysis({
      actorHash,
      authoritativeTaskGraph: analysisFile.authoritativeTaskGraph,
      bytes: analysisFile.bytes,
      extension: analysisFile.extension,
      fileHash,
      fileName: analysisFile.fileName,
      mimeType: analysisFile.mimeType,
      sourceNamespace: draft?.sourceNamespace ?? undefined,
      sourceType: authoritativeSourceType,
      workspaceId: context.workspace.id,
    });

    return jsonResponse({
      analysis: draft,
      fileHash,
      responseId: response.id,
      status: "queued",
    });
  } catch (error) {
    const failureMessage = getAIAnalysisFailureMessage(
      error,
      "AI mapping suggestions could not be generated, so the deterministic mapping is shown for review.",
    );
    if (draft) {
      return jsonResponse({
        analysis: {
          ...draft,
          warnings: [...draft.warnings, failureMessage],
        },
        fileHash,
        responseId: null,
        status: "completed",
      });
    }
    return textResponse(
      getAIAnalysisFailureMessage(
        error,
        "The file could not be queued for AI analysis",
      ),
      502,
    );
  }
}

export async function GET(request: Request): Promise<Response> {
  const session = await auth();
  if (!session?.user) return textResponse("Unauthorized", 401);

  const url = new URL(request.url);
  const parsed = pollQuerySchema.safeParse({
    responseId: url.searchParams.get("responseId"),
    workspaceSlug: url.searchParams.get("workspaceSlug"),
    fileHash: url.searchParams.get("fileHash"),
  });
  if (!parsed.success) {
    return textResponse("Invalid import analysis request", 400);
  }

  const context = await getWorkspaceContext(parsed.data.workspaceSlug, session);
  if (!context.ok) return context.error;
  if (!process.env.OPENAI_API_KEY) {
    return textResponse("AI file analysis is not configured", 503);
  }

  return pollBackgroundAnalysis({
    actorId: context.session.user.id,
    fileHash: parsed.data.fileHash,
    responseId: parsed.data.responseId,
    workspaceId: context.workspace.id,
  });
}
