/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  MUTATING_ACTION_TOOL_NAMES,
  MUTATING_TOOL_ACTIONS,
  MUTATION_TOOL_NAMES,
} from "@/lib/ai/tool-policy";
import {
  getMayaToolInvalidationKeys,
  MAYA_TOOL_INVALIDATION_PROFILES,
} from "./tool-query-invalidation";

const WORKSPACE_SLUG = "acme";

const getKeys = (toolName: string, input: unknown = {}) =>
  getMayaToolInvalidationKeys({
    input,
    toolName,
    workspaceSlug: WORKSPACE_SLUG,
  });

describe("Maya tool query invalidation", () => {
  it("classifies every mutation-capable tool with an invalidation profile", () => {
    const mutationCapableTools = new Set([
      ...MUTATION_TOOL_NAMES,
      ...MUTATING_ACTION_TOOL_NAMES,
    ]);

    expect(Object.keys(MAYA_TOOL_INVALIDATION_PROFILES).sort()).toEqual(
      Array.from(mutationCapableTools).sort(),
    );
  });

  it.each(
    MUTATION_TOOL_NAMES.filter(
      (toolName) => toolName !== "createGitHubInstallSessionTool",
    ),
  )("invalidates cached data after %s", (toolName) => {
    expect(getKeys(toolName)).not.toHaveLength(0);
  });

  it.each(
    Object.entries(MUTATING_TOOL_ACTIONS).flatMap(([toolName, actions]) =>
      Array.from(actions).map((action) => [toolName, action] as const),
    ),
  )("invalidates cached data after %s:%s", (toolName, action) => {
    expect(getKeys(toolName, { action })).not.toHaveLength(0);
  });

  it("does not invalidate data for read actions or install-link creation", () => {
    expect(getKeys("comments", { action: "list-comments" })).toEqual([]);
    expect(getKeys("labels", { action: "list-labels" })).toEqual([]);
    expect(getKeys("createGitHubInstallSessionTool")).toEqual([]);
  });

  it("refreshes story totals, analytics, and calendars after story lifecycle changes", () => {
    expect(getKeys("bulkCreateStories")).toEqual(
      expect.arrayContaining([
        ["stories", WORKSPACE_SLUG],
        ["totalStories", WORKSPACE_SLUG],
        ["calendar", WORKSPACE_SLUG],
        ["analytics", WORKSPACE_SLUG],
      ]),
    );
  });

  it("refreshes the exact story-link cache when its story is known", () => {
    expect(
      getKeys("links", { action: "add-link", storyId: "story-1" }),
    ).toEqual(
      expect.arrayContaining([
        ["stories", WORKSPACE_SLUG],
        ["story-links", "story-1"],
      ]),
    );
  });

  it("refreshes broad GitHub caches for entity-specific GitHub mutations", () => {
    expect(getKeys("updateGitHubTeamSettingsTool")).toContainEqual([
      "github",
      WORKSPACE_SLUG,
    ]);
    expect(getKeys("postRequestGitHubCommentTool")).toEqual(
      expect.arrayContaining([
        ["github", WORKSPACE_SLUG],
        ["integration-requests", WORKSPACE_SLUG],
      ]),
    );
  });
});
