import { asSchema, jsonSchema } from "ai";
import type { FlexibleSchema, ToolSet } from "ai";
import { compactToolOutput } from "./compact-tool-output";
import { requiresMutationApproval } from "./tool-policy";

type ValidationIssue = {
  path?: PropertyKey[];
};

const getValidationIssues = (error: Error): ValidationIssue[] => {
  const candidates: unknown[] = [error];
  const visited = new Set<unknown>();

  while (candidates.length > 0) {
    const candidate = candidates.shift();
    if (!candidate || typeof candidate !== "object" || visited.has(candidate)) {
      continue;
    }
    visited.add(candidate);

    if ("issues" in candidate && Array.isArray(candidate.issues)) {
      return candidate.issues as ValidationIssue[];
    }
    if ("cause" in candidate) candidates.push(candidate.cause);
  }

  return [];
};

const getValueAtPath = (value: unknown, path: PropertyKey[]) => {
  let current = value;
  for (const segment of path) {
    if (!current || typeof current !== "object") return undefined;
    current = (current as Record<PropertyKey, unknown>)[segment];
  }
  return current;
};

const deleteObjectPropertyAtPath = (value: unknown, path: PropertyKey[]) => {
  if (path.length === 0) return false;

  let parent = value;
  for (const segment of path.slice(0, -1)) {
    if (!parent || typeof parent !== "object") return false;
    parent = (parent as Record<PropertyKey, unknown>)[segment];
  }
  if (!parent || typeof parent !== "object") return false;

  const property = path.at(-1);
  if (typeof property !== "string") return false;
  if (!Object.hasOwn(parent, property)) return false;

  return Reflect.deleteProperty(parent, property);
};

/**
 * OpenAI strict tool schemas encode optional properties as required nullable
 * properties. Some calls therefore contain null for an omitted non-nullable
 * field. Remove only nulls that the source validator identifies as invalid;
 * nulls accepted by the schema remain intact for intentional clear operations.
 */
export const validateToolInputWithStrictNullNormalization = async (
  inputSchema: FlexibleSchema<unknown>,
  input: unknown,
) => {
  const schema = asSchema(inputSchema);
  if (!schema.validate) {
    return { success: true as const, value: input };
  }

  let candidate = input;
  for (let attempt = 0; attempt < 8; attempt += 1) {
    // eslint-disable-next-line no-await-in-loop -- Each pass removes only null paths rejected by the source schema.
    const result = await schema.validate(candidate);
    if (result.success) return result;

    const currentCandidate = candidate;
    const invalidNullPaths: PropertyKey[][] = [];
    for (const issue of getValidationIssues(result.error)) {
      if (
        Array.isArray(issue.path) &&
        issue.path.length > 0 &&
        getValueAtPath(currentCandidate, issue.path) === null
      ) {
        invalidNullPaths.push(issue.path);
      }
    }
    if (invalidNullPaths.length === 0) return result;

    const nextCandidate = structuredClone(currentCandidate);
    let removedAny = false;
    for (const path of invalidNullPaths) {
      removedAny =
        deleteObjectPropertyAtPath(nextCandidate, path) || removedAny;
    }
    if (!removedAny) return result;
    candidate = nextCandidate;
  }

  return schema.validate(candidate);
};

const withStrictNullNormalization = (inputSchema: FlexibleSchema<unknown>) => {
  const schema = asSchema(inputSchema);
  return jsonSchema(schema.jsonSchema, {
    validate: (input) =>
      validateToolInputWithStrictNullNormalization(inputSchema, input),
  });
};

export const withCompactModelOutputs = <TOOLS extends ToolSet>(
  toolSet: TOOLS,
): TOOLS =>
  Object.fromEntries(
    Object.entries(toolSet).map(([name, tool]) => [
      name,
      {
        ...tool,
        inputSchema: withStrictNullNormalization(tool.inputSchema),
        needsApproval: (input: unknown) =>
          requiresMutationApproval(name, input),
        ...(tool.toModelOutput
          ? {}
          : {
              toModelOutput: ({ output }: { output: unknown }) => ({
                type: "json" as const,
                value: compactToolOutput(output, { toolName: name }),
              }),
            }),
      },
    ]),
  ) as unknown as TOOLS;
