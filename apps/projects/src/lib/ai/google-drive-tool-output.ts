import type { UIMessage } from "ai";
import {
  canMayaReadGoogleDriveFile,
  isTrustedGoogleDriveWebViewLink,
} from "@/modules/google-drive/capabilities";

const GOOGLE_DRIVE_CONTENT_TOOL_NAME =
  "getLinkedGoogleFileContentTool" as const;
const GOOGLE_DRIVE_CONTENT_TOOL_PART_TYPE =
  `tool-${GOOGLE_DRIVE_CONTENT_TOOL_NAME}` as const;
const MODEL_ONLY_CONTENT = Symbol("google-drive-model-only-content");

type GoogleDriveContentFile = {
  name: string;
  mimeType: string;
  webViewLink: string;
  modifiedTime?: string;
  content: string;
  contentType: "text/plain" | "text/markdown" | "text/csv";
  contentTruncated: boolean;
  bytesRead: number;
};

type GoogleDriveContentReceipt = {
  success: true;
  file: {
    name: string;
    mimeType: string;
    webViewLink: string;
    modifiedTime?: string;
    contentType: GoogleDriveContentFile["contentType"];
    contentTruncated: boolean;
    bytesRead: number;
    contentRetained: false;
    untrustedExternalContent: true;
  };
};

type GoogleDriveContentFailure = {
  success: false;
  error: string;
};

type GoogleDriveContentExecutionOutput = GoogleDriveContentReceipt & {
  [MODEL_ONLY_CONTENT]?: string;
};

const asRecord = (value: unknown): Record<PropertyKey, unknown> | undefined =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<PropertyKey, unknown>)
    : undefined;

const boundedError = (value: unknown) =>
  typeof value === "string" && value.trim()
    ? value.trim().slice(0, 500)
    : "Maya could not read the selected Google Drive file.";

/**
 * Return the only Google Drive read result that may cross a client or storage
 * boundary. Raw content is deliberately excluded.
 */
export const getGoogleDriveContentReceipt = (
  output: unknown,
): GoogleDriveContentFailure | GoogleDriveContentReceipt => {
  const result = asRecord(output);
  if (!result || result.success !== true) {
    return {
      success: false,
      error: boundedError(result?.error),
    };
  }

  const file = asRecord(result.file);
  if (
    !file ||
    typeof file.name !== "string" ||
    typeof file.mimeType !== "string" ||
    !canMayaReadGoogleDriveFile(file.mimeType) ||
    typeof file.webViewLink !== "string" ||
    !isTrustedGoogleDriveWebViewLink(file.webViewLink) ||
    (file.modifiedTime !== undefined &&
      typeof file.modifiedTime !== "string") ||
    (file.contentType !== "text/plain" &&
      file.contentType !== "text/markdown" &&
      file.contentType !== "text/csv") ||
    typeof file.contentTruncated !== "boolean" ||
    typeof file.bytesRead !== "number" ||
    !Number.isFinite(file.bytesRead) ||
    file.bytesRead < 0
  ) {
    return {
      success: false,
      error: "Google Drive content metadata was invalid.",
    };
  }

  return {
    success: true,
    file: {
      name: file.name.slice(0, 500),
      mimeType: file.mimeType.slice(0, 255),
      webViewLink: file.webViewLink.slice(0, 2_048),
      ...(file.modifiedTime
        ? { modifiedTime: file.modifiedTime.slice(0, 100) }
        : {}),
      contentType: file.contentType,
      contentTruncated: file.contentTruncated,
      bytesRead: Math.max(0, Math.trunc(file.bytesRead)),
      contentRetained: false,
      untrustedExternalContent: true,
    },
  };
};

/**
 * Keep Drive text in-memory only long enough to create the next model step.
 * Symbols and non-enumerable properties are omitted by JSON serialization,
 * while the explicit transcript redactor below removes them before saving.
 */
export const createGoogleDriveContentExecutionOutput = (
  file: GoogleDriveContentFile,
): GoogleDriveContentExecutionOutput => {
  const receipt = getGoogleDriveContentReceipt({
    success: true,
    file,
  });
  if (!receipt.success) {
    throw new Error(receipt.error);
  }

  Object.defineProperty(receipt, MODEL_ONLY_CONTENT, {
    configurable: false,
    enumerable: false,
    value: file.content,
    writable: false,
  });

  return receipt;
};

/**
 * AI SDK invokes this projection directly after execute(). Only that in-memory
 * object contains MODEL_ONLY_CONTENT. Replayed/persisted receipts therefore do
 * not restore content and tell the model to require a fresh selection/read.
 */
export const toGoogleDriveContentModelOutput = ({
  output,
}: {
  output: unknown;
}) => {
  const receipt = getGoogleDriveContentReceipt(output);
  const result = asRecord(output);
  const modelOnlyContent = result?.[MODEL_ONLY_CONTENT];

  if (!receipt.success) {
    return { type: "json" as const, value: receipt };
  }

  return {
    type: "json" as const,
    value: {
      ...receipt,
      file: {
        ...receipt.file,
        ...(typeof modelOnlyContent === "string"
          ? {
              content: modelOnlyContent,
              contentAvailableForCurrentResponse: true,
            }
          : {
              contentAvailableForCurrentResponse: false,
              contentUnavailableReason:
                "Google Drive content is not retained in chat history. The user must explicitly select the file again before it can be read on a later turn.",
            }),
      },
    },
  };
};

/**
 * Defense in depth for legacy/crafted transcripts and for onFinish output.
 * This function replaces every completed Drive read result with its safe
 * metadata receipt before model-history conversion or transcript persistence.
 */
export const redactGoogleDriveContentFromMessages = <Message extends UIMessage>(
  messages: Message[],
): Message[] =>
  messages.map((message) => {
    const hasGoogleDriveContent = message.parts.some(
      (part) =>
        part.type === GOOGLE_DRIVE_CONTENT_TOOL_PART_TYPE && "output" in part,
    );
    if (!hasGoogleDriveContent) return message;

    return {
      ...message,
      parts: message.parts.map((part) =>
        part.type === GOOGLE_DRIVE_CONTENT_TOOL_PART_TYPE && "output" in part
          ? ({
              ...part,
              output: getGoogleDriveContentReceipt(part.output),
            } as typeof part)
          : part,
      ),
    } as Message;
  });
