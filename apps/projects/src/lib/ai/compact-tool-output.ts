import type { JSONValue } from "ai";
import { getGoogleDriveContentReceipt } from "./google-drive-tool-output";

const DEFAULT_LIMITS = {
  arrayItems: 20,
  depth: 5,
  stringCharacters: 1200,
} as const;

const STRICT_LIMITS = {
  arrayItems: 8,
  depth: 4,
  stringCharacters: 500,
} as const;

const MAX_MODEL_OUTPUT_CHARACTERS = 24_000;
const MAX_BULK_STORY_RECEIPTS = 50;
const MAX_RECEIPT_TITLE_CHARACTERS = 240;
const MAX_DOCUMENT_CONTENT_CHARACTERS = 20_000;
const GITHUB_INSTALL_SESSION_TOOL = "createGitHubInstallSessionTool";
const GOOGLE_DRIVE_CONTENT_TOOL = "getLinkedGoogleFileContentTool";

const ANALYTICS_PAYLOAD_KEYS: Readonly<Record<string, string>> = {
  objectiveProgressReportTool: "progress",
  pulseReportTool: "report",
  sprintPerformanceReportTool: "analytics",
  storyPerformanceReportTool: "analytics",
  teamPerformanceReportTool: "performance",
  timelineTrendsReportTool: "trends",
  workspaceCommandCenterReportTool: "report",
  workspacePerformanceReportTool: "overview",
};

const ANALYTICS_LIMITS = {
  arrayItems: 5,
  depth: 8,
  stringCharacters: 500,
} as const;

const STRICT_ANALYTICS_LIMITS = {
  arrayItems: 2,
  depth: 7,
  stringCharacters: 240,
} as const;

const ATTACHMENT_STORAGE_URL_KEYS = new Set([
  "downloadUrl",
  "signedUrl",
  "storageUrl",
  "url",
]);

const OMITTED_MODEL_KEYS = new Set([
  "avatar",
  "avatarUrl",
  "contentHTML",
  "descriptionHTML",
  "email",
  "html",
  "image",
  "imageUrl",
  "picture",
  "profilePicture",
]);

type CompactLimits = {
  arrayItems: number;
  depth: number;
  stringCharacters: number;
};

type CompactToolOutputOptions = {
  toolName?: string;
};

const compactString = (value: string, maxCharacters: number) => {
  if (/^data:[^;]+;base64,/i.test(value)) return "[binary data omitted]";
  if (value.length <= maxCharacters) return value;

  return `${value.slice(0, maxCharacters).trimEnd()}…`;
};

const compactValue = (
  value: unknown,
  limits: CompactLimits,
  depth = 0,
): JSONValue => {
  if (value === null || value === undefined) return null;
  if (typeof value === "string") {
    return compactString(value, limits.stringCharacters);
  }
  if (typeof value === "number" || typeof value === "boolean") return value;
  if (typeof value === "bigint") return value.toString();
  if (depth >= limits.depth) return "[nested data omitted]";

  if (Array.isArray(value)) {
    return value
      .slice(0, limits.arrayItems)
      .map((item) => compactValue(item, limits, depth + 1));
  }

  if (typeof value !== "object") return String(value);

  const compacted: Record<string, JSONValue> = {};
  const omittedItems: Record<string, number> = {};

  for (const [key, item] of Object.entries(value)) {
    if (item === undefined || OMITTED_MODEL_KEYS.has(key)) continue;

    compacted[key] = compactValue(item, limits, depth + 1);
    if (Array.isArray(item) && item.length > limits.arrayItems) {
      omittedItems[key] = item.length - limits.arrayItems;
    }
  }

  if (Object.keys(omittedItems).length > 0) {
    compacted.modelItemsOmitted = omittedItems;
  }

  return compacted;
};

const asRecord = (value: unknown): Record<string, unknown> | undefined =>
  value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;

const getGitHubInstallSessionModelOutput = (
  value: unknown,
  toolName?: string,
): JSONValue | undefined => {
  if (toolName !== GITHUB_INSTALL_SESSION_TOOL) return undefined;

  const source = asRecord(value);
  if (!source) {
    return {
      success: false,
      error: "GitHub install session could not be created.",
    };
  }

  if (source.success === true) {
    return {
      success: true,
      installSessionReady: typeof source.installUrl === "string",
      message:
        "GitHub install session created. Continue using the link shown in the interface.",
    };
  }

  const error = typeof source.error === "string" ? source.error : "";
  return {
    success: false,
    error: error.startsWith("Authentication required")
      ? "Authentication required to create a GitHub install session."
      : "GitHub install session could not be created.",
  };
};

const isStoryReceipt = (value: unknown) => {
  const record = asRecord(value);
  return (
    Boolean(record) &&
    typeof record?.id === "string" &&
    typeof record.title === "string"
  );
};

const isBulkCreateReceipt = (
  value: Record<string, unknown>,
  toolName?: string,
) =>
  toolName === "bulkCreateStories" ||
  (typeof value.createdCount === "number" &&
    Array.isArray(value.stories) &&
    value.stories.every(isStoryReceipt));

const isBulkDeleteReceipt = (
  value: Record<string, unknown>,
  toolName?: string,
) =>
  toolName === "bulkDeleteStories" ||
  (typeof value.deletedCount === "number" &&
    typeof value.requestedCount === "number" &&
    Array.isArray(value.storyIds));

const isDocumentDetailsOutput = (
  value: Record<string, unknown>,
  toolName?: string,
) => {
  if (toolName === "getDocumentDetailsTool") return true;

  const document = asRecord(value.document);
  return (
    typeof document?.id === "string" &&
    typeof document.title === "string" &&
    typeof document.content === "string"
  );
};

const isAttachmentCollectionOutput = (
  value: Record<string, unknown>,
  toolName?: string,
) => {
  if (toolName === "listAttachments") return true;
  if (!Array.isArray(value.attachments)) return false;

  return value.attachments.every((attachment) => {
    const record = asRecord(attachment);
    return (
      typeof record?.id === "string" &&
      typeof record.filename === "string" &&
      (typeof record.mimeType === "string" || typeof record.size === "number")
    );
  });
};

const updateOmittedItemCount = ({
  compacted,
  key,
  omittedCount,
}: {
  compacted: Record<string, JSONValue>;
  key: string;
  omittedCount: number;
}) => {
  const existing = (asRecord(compacted.modelItemsOmitted) ?? {}) as Record<
    string,
    JSONValue
  >;
  const next: Record<string, JSONValue> =
    omittedCount > 0
      ? { ...existing, [key]: omittedCount }
      : Object.fromEntries(
          Object.entries(existing).filter(
            ([existingKey]) => existingKey !== key,
          ),
        );

  if (Object.keys(next).length > 0) compacted.modelItemsOmitted = next;
  else delete compacted.modelItemsOmitted;
};

const preserveBulkMutationReceipts = (
  value: unknown,
  compacted: JSONValue,
  limits: CompactLimits,
  toolName?: string,
): JSONValue => {
  const source = asRecord(value);
  const compactedRecord = asRecord(compacted) as
    | Record<string, JSONValue>
    | undefined;
  if (!source || !compactedRecord) return compacted;

  if (isBulkCreateReceipt(source, toolName) && Array.isArray(source.stories)) {
    const stories = source.stories.filter(isStoryReceipt);
    compactedRecord.stories = stories
      .slice(0, MAX_BULK_STORY_RECEIPTS)
      .map((story) => {
        const receipt = story as { id: string; title: string };
        return {
          id: compactString(receipt.id, limits.stringCharacters),
          title: compactString(
            receipt.title,
            Math.min(limits.stringCharacters, MAX_RECEIPT_TITLE_CHARACTERS),
          ),
        };
      });
    updateOmittedItemCount({
      compacted: compactedRecord,
      key: "stories",
      omittedCount: Math.max(stories.length - MAX_BULK_STORY_RECEIPTS, 0),
    });
  }

  if (isBulkDeleteReceipt(source, toolName)) {
    for (const key of ["storyIds", "missingStoryIds"] as const) {
      if (!Array.isArray(source[key])) continue;

      const storyIds = source[key].filter(
        (storyId): storyId is string => typeof storyId === "string",
      );
      compactedRecord[key] = storyIds
        .slice(0, MAX_BULK_STORY_RECEIPTS)
        .map((storyId) => compactString(storyId, limits.stringCharacters));
      updateOmittedItemCount({
        compacted: compactedRecord,
        key,
        omittedCount: Math.max(storyIds.length - MAX_BULK_STORY_RECEIPTS, 0),
      });
    }
  }

  return compactedRecord;
};

const preserveDocumentContent = (
  value: unknown,
  compacted: JSONValue,
  toolName?: string,
): JSONValue => {
  const source = asRecord(value);
  const compactedRecord = asRecord(compacted) as
    | Record<string, JSONValue>
    | undefined;
  if (
    !source ||
    !compactedRecord ||
    !isDocumentDetailsOutput(source, toolName)
  ) {
    return compacted;
  }

  const sourceDocument = asRecord(source.document);
  const compactedDocument = asRecord(compactedRecord.document) as
    | Record<string, JSONValue>
    | undefined;
  if (
    !sourceDocument ||
    !compactedDocument ||
    typeof sourceDocument.content !== "string"
  ) {
    return compacted;
  }

  compactedDocument.content = compactString(
    sourceDocument.content,
    MAX_DOCUMENT_CONTENT_CHARACTERS,
  );
  compactedDocument.contentTruncated =
    sourceDocument.contentTruncated === true ||
    sourceDocument.content.length > MAX_DOCUMENT_CONTENT_CHARACTERS;

  return compactedRecord;
};

const redactAttachmentStorageUrls = (
  value: unknown,
  compacted: JSONValue,
  toolName?: string,
): JSONValue => {
  const source = asRecord(value);
  const compactedRecord = asRecord(compacted) as
    | Record<string, JSONValue>
    | undefined;
  if (
    !source ||
    !compactedRecord ||
    !isAttachmentCollectionOutput(source, toolName) ||
    !Array.isArray(compactedRecord.attachments)
  ) {
    return compacted;
  }

  compactedRecord.attachments = compactedRecord.attachments.map(
    (attachment) => {
      const compactedAttachment = asRecord(attachment);
      if (!compactedAttachment) return attachment;

      return Object.fromEntries(
        Object.entries(compactedAttachment).filter(
          ([key]) => !ATTACHMENT_STORAGE_URL_KEYS.has(key),
        ),
      ) as Record<string, JSONValue>;
    },
  );

  return compactedRecord;
};

const compactWithToolData = (
  value: unknown,
  limits: CompactLimits,
  toolName?: string,
) => {
  const withReceipts = preserveBulkMutationReceipts(
    value,
    compactValue(value, limits),
    limits,
    toolName,
  );

  const withDocumentContent = preserveDocumentContent(
    value,
    withReceipts,
    toolName,
  );

  return redactAttachmentStorageUrls(value, withDocumentContent, toolName);
};

const getMinimalDocumentOutput = (
  value: unknown,
  toolName?: string,
): JSONValue | undefined => {
  const source = asRecord(value);
  if (!source || !isDocumentDetailsOutput(source, toolName)) return undefined;

  const document = asRecord(source.document);
  if (!document || typeof document.content !== "string") return undefined;

  const minimalDocument: Record<string, JSONValue> = {
    content: compactString(document.content, MAX_DOCUMENT_CONTENT_CHARACTERS),
    contentTruncated:
      document.contentTruncated === true ||
      document.content.length > MAX_DOCUMENT_CONTENT_CHARACTERS,
  };

  for (const key of ["id", "title", "visibility", "relatedWorkCount"]) {
    const item = document[key];
    if (
      typeof item === "string" ||
      typeof item === "number" ||
      typeof item === "boolean"
    ) {
      minimalDocument[key] = compactValue(item, STRICT_LIMITS);
    }
  }

  return {
    success: source.success === true,
    document: minimalDocument,
    modelOutputTruncated: true,
  };
};

const getAnalyticsPayloadKey = (
  source: Record<string, unknown>,
  toolName?: string,
) => {
  if (toolName && ANALYTICS_PAYLOAD_KEYS[toolName]) {
    return ANALYTICS_PAYLOAD_KEYS[toolName];
  }

  const kind = typeof source.kind === "string" ? source.kind : "";
  if (!kind.includes("report") && !kind.includes("analytics")) return undefined;

  return [
    "report",
    "overview",
    "analytics",
    "progress",
    "performance",
    "trends",
  ].find((key) => source[key] !== undefined);
};

const getAnalyticsProjection = (
  value: unknown,
  toolName?: string,
): JSONValue | undefined => {
  const source = asRecord(value);
  if (!source) return undefined;

  const payloadKey = getAnalyticsPayloadKey(source, toolName);
  if (!payloadKey) return undefined;

  const buildProjection = (limits: CompactLimits) => {
    const projection: Record<string, JSONValue> = {
      modelOutputTruncated: true,
    };

    for (const key of [
      "success",
      "kind",
      "title",
      "message",
      "error",
      "filters",
      "userId",
      "focusMember",
    ]) {
      if (source[key] !== undefined) {
        projection[key] = compactValue(source[key], limits);
      }
    }

    projection[payloadKey] = compactValue(source[payloadKey], limits);
    return projection;
  };

  const projection = buildProjection(ANALYTICS_LIMITS);
  if (JSON.stringify(projection).length <= MAX_MODEL_OUTPUT_CHARACTERS) {
    return projection;
  }

  return buildProjection(STRICT_ANALYTICS_LIMITS);
};

const getMinimalOutput = (value: unknown): JSONValue => {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return compactValue(value, STRICT_LIMITS);
  }

  const record = value as Record<string, unknown>;
  const minimalOutput: Record<string, JSONValue> = {
    modelOutputTruncated: true,
  };

  for (const key of [
    "success",
    "kind",
    "message",
    "error",
    "count",
    "totalCount",
    "returnedCount",
  ]) {
    const item = record[key];
    if (
      typeof item === "string" ||
      typeof item === "number" ||
      typeof item === "boolean"
    ) {
      minimalOutput[key] = compactValue(item, STRICT_LIMITS);
    }
  }

  return minimalOutput;
};

export const compactToolOutput = (
  value: unknown,
  { toolName }: CompactToolOutputOptions = {},
): JSONValue => {
  if (toolName === GOOGLE_DRIVE_CONTENT_TOOL) {
    return getGoogleDriveContentReceipt(value);
  }

  const secureToolOutput = getGitHubInstallSessionModelOutput(value, toolName);
  if (secureToolOutput) return secureToolOutput;

  const compacted = compactWithToolData(value, DEFAULT_LIMITS, toolName);
  if (JSON.stringify(compacted).length <= MAX_MODEL_OUTPUT_CHARACTERS) {
    return compacted;
  }

  const strictCompacted = compactWithToolData(value, STRICT_LIMITS, toolName);
  if (JSON.stringify(strictCompacted).length <= MAX_MODEL_OUTPUT_CHARACTERS) {
    return strictCompacted;
  }

  const minimalDocumentOutput = getMinimalDocumentOutput(value, toolName);
  if (
    minimalDocumentOutput &&
    JSON.stringify(minimalDocumentOutput).length <= MAX_MODEL_OUTPUT_CHARACTERS
  ) {
    return minimalDocumentOutput;
  }

  const analyticsProjection = getAnalyticsProjection(value, toolName);
  if (
    analyticsProjection &&
    JSON.stringify(analyticsProjection).length <= MAX_MODEL_OUTPUT_CHARACTERS
  ) {
    return analyticsProjection;
  }

  return getMinimalOutput(value);
};
