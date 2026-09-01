import { createHash } from "node:crypto";

const MAYA_PROMPT_CACHE_NAMESPACE = "maya-v4-tool-search";
const WORKSPACE_DIGEST_LENGTH = 32;

/**
 * OpenAI limits prompt cache keys to 64 characters. Hashing the workspace ID
 * preserves stable per-workspace cache isolation without coupling correctness
 * to the length of either the namespace or the persisted identifier.
 */
export const getMayaPromptCacheKey = (workspaceId: string) => {
  const workspaceDigest = createHash("sha256")
    .update(workspaceId)
    .digest("hex")
    .slice(0, WORKSPACE_DIGEST_LENGTH);

  return `${MAYA_PROMPT_CACHE_NAMESPACE}:${workspaceDigest}`;
};
