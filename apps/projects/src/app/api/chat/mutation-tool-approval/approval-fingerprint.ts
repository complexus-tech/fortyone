import "server-only";

import { createHash } from "node:crypto";
import type { MayaToolName } from "@/lib/ai/tool-policy";

export type MutationToolApproval = {
  approved: boolean;
  input: unknown;
  toolCallId: string;
  toolName: MayaToolName;
};

const normalizeForFingerprint = (value: unknown): unknown => {
  if (Array.isArray(value)) return value.map(normalizeForFingerprint);
  if (!value || typeof value !== "object") return value;

  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, item]) => [key, normalizeForFingerprint(item)]),
  );
};

const getFingerprint = (value: unknown) =>
  createHash("sha256")
    .update(JSON.stringify(normalizeForFingerprint(value)))
    .digest("hex");

export const getApprovalFingerprint = (approval: MutationToolApproval) =>
  getFingerprint({
    approved: approval.approved,
    input: approval.input,
    toolName: approval.toolName,
  });

export const getPreparedApprovalFingerprint = (
  approval: MutationToolApproval,
  input: unknown,
) =>
  getFingerprint({
    approved: true,
    input,
    toolName: approval.toolName,
  });
