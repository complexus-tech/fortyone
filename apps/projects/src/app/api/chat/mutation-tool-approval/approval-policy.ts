import "server-only";

import type { FlexibleSchema, ToolSet } from "ai";
import { validateToolInputWithStrictNullNormalization } from "@/lib/ai/model-tools";
import { tools } from "@/lib/ai/tools";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import {
  type MayaToolName,
  requiresMutationApproval,
  toApprovedMutationInput,
} from "@/lib/ai/tool-policy";
import type { MutationToolApproval } from "./approval-fingerprint";

export type { MutationToolApproval } from "./approval-fingerprint";

export type PreparedApprovedMutation = {
  execute: NonNullable<ToolSet[string]["execute"]>;
  input: unknown;
};

const isRegisteredToolName = (toolName: string): toolName is MayaToolName =>
  Object.hasOwn(tools, toolName);

export const getMutationToolApprovals = (
  messages: MayaUIMessage[],
): MutationToolApproval[] => {
  const lastMessage = messages.at(-1);
  if (lastMessage?.role !== "assistant") return [];

  const approvals: MutationToolApproval[] = [];
  const seenToolCallIds = new Set<string>();

  for (const part of lastMessage.parts) {
    if (!part.type.startsWith("tool-") || !("state" in part)) continue;
    if (part.state !== "approval-responded" || !("approval" in part)) {
      continue;
    }
    if (!("input" in part) || !("toolCallId" in part)) continue;
    if (typeof part.toolCallId !== "string") continue;
    if (seenToolCallIds.has(part.toolCallId)) continue;

    const toolName = part.type.slice("tool-".length);
    const approved = part.approval.approved;
    if (
      typeof approved !== "boolean" ||
      !isRegisteredToolName(toolName) ||
      !requiresMutationApproval(toolName, part.input)
    ) {
      continue;
    }

    seenToolCallIds.add(part.toolCallId);
    approvals.push({
      approved,
      input: part.input,
      toolCallId: part.toolCallId,
      toolName,
    });
  }

  return approvals;
};

const validateToolInput = async (toolName: MayaToolName, input: unknown) => {
  const registeredTool = tools[toolName];
  const validation = await validateToolInputWithStrictNullNormalization(
    registeredTool.inputSchema as FlexibleSchema<unknown>,
    input,
  );
  if (!validation.success) throw validation.error;
  if (!requiresMutationApproval(toolName, validation.value)) {
    throw new Error(`The ${toolName} call is not a mutation.`);
  }

  return validation.value;
};

export const prepareApprovedMutation = async (
  approval: MutationToolApproval,
): Promise<PreparedApprovedMutation> => {
  const registeredTool = tools[approval.toolName] as ToolSet[string];
  if (!registeredTool.execute) {
    throw new Error(`The ${approval.toolName} tool cannot be executed.`);
  }

  const validatedInput = await validateToolInput(
    approval.toolName,
    approval.input,
  );
  return {
    execute: registeredTool.execute,
    input: toApprovedMutationInput(approval.toolName, validatedInput),
  };
};
