import assert from "node:assert/strict";
import test from "node:test";
import { buildMayaLegacyContext } from "./request-context.ts";

const input: Parameters<typeof buildMayaLegacyContext>[0] = {
  scope: { userId: "user-one", workspace: "my-workspace", sessionEpoch: 1 },
  workspace: {
    id: "workspace-one",
    name: "My workspace",
    slug: "my-workspace",
    userRole: "member",
    isActive: true,
    deletedAt: null,
    color: "grey",
    avatarUrl: null,
    trialEndsOn: null,
    createdAt: "2026-01-01",
    updatedAt: "2026-01-01",
  },
  profile: { id: "user-one", username: "person" },
  settings: {
    storyTerm: "story",
    sprintTerm: "cycle",
    objectiveTerm: "project",
    keyResultTerm: "focus area",
    objectiveEnabled: true,
    keyResultEnabled: true,
  },
  memories: [
    {
      id: "memory-one",
      workspaceId: "workspace-one",
      userId: "user-one",
      content: "Keep responses concise.",
    },
  ],
  subscription: null,
};

test("legacy request carries the exact workspace context and configured plural terms", () => {
  const context = buildMayaLegacyContext(input);
  assert.deepEqual(context.workspace, input.workspace);
  assert.deepEqual(context.memories, input.memories);
  assert.equal(context.username, input.profile.username);
  assert.deepEqual(context.terminology, {
    stories: "stories",
    sprints: "cycles",
    objectives: "projects",
    keyResults: "focus areas",
  });
  assert.equal("subscription" in context, false);
  assert.equal("totalMessages" in context, false);
});

test("context from another account or workspace cannot be sent", () => {
  for (const changed of [
    { profile: { ...input.profile, id: "other-user" } },
    { workspace: { ...input.workspace, slug: "other-workspace" } },
    { workspace: { ...input.workspace, isActive: false } },
    { memories: [{ ...input.memories[0], userId: "other-user" }] },
    { memories: [{ ...input.memories[0], workspaceId: "other-workspace" }] },
  ])
    assert.throws(
      () => buildMayaLegacyContext({ ...input, ...changed }),
      /context changed/,
    );
});

test("missing terminology is surfaced rather than silently replacing workspace terms", () => {
  assert.throws(
    () =>
      buildMayaLegacyContext({
        ...input,
        settings: { ...input.settings, storyTerm: "" as never },
      }),
    /terminology could not be loaded/,
  );
});
