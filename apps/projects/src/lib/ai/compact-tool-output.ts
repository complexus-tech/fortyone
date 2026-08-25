import type { JSONValue } from "ai";

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

const OMITTED_MODEL_KEYS = new Set([
  "avatar",
  "avatarUrl",
  "contentHTML",
  "descriptionHTML",
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

export const compactToolOutput = (value: unknown): JSONValue => {
  const compacted = compactValue(value, DEFAULT_LIMITS);
  if (JSON.stringify(compacted).length <= MAX_MODEL_OUTPUT_CHARACTERS) {
    return compacted;
  }

  const strictCompacted = compactValue(value, STRICT_LIMITS);
  return JSON.stringify(strictCompacted).length <= MAX_MODEL_OUTPUT_CHARACTERS
    ? strictCompacted
    : getMinimalOutput(value);
};
