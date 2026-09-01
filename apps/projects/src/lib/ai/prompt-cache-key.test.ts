/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this docblock.

import { getMayaPromptCacheKey } from "./prompt-cache-key";

describe("getMayaPromptCacheKey", () => {
  it("returns a stable provider-safe key", () => {
    const workspaceId = "00000000-0000-0000-0000-000000000000";
    const key = getMayaPromptCacheKey(workspaceId);

    expect(key).toBe(getMayaPromptCacheKey(workspaceId));
    expect(key.length).toBeLessThanOrEqual(64);
    expect(key).toMatch(/^maya-v4-tool-search:[a-f0-9]{32}$/);
  });

  it("isolates cache entries between workspaces", () => {
    expect(getMayaPromptCacheKey("workspace-a")).not.toBe(
      getMayaPromptCacheKey("workspace-b"),
    );
  });

  it("remains bounded for unexpectedly long workspace identifiers", () => {
    expect(getMayaPromptCacheKey("workspace".repeat(1_000)).length).toBeLessThanOrEqual(
      64,
    );
  });
});
