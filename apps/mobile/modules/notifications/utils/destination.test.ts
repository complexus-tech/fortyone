import assert from "node:assert/strict";
import test from "node:test";
import {
  notificationWebURL,
  resolveNotificationDestination,
} from "./destination.ts";

const context = {
  applicationURL: "https://cloud.fortyone.app",
  workspace: "acme",
  loadObjective: async (id: string) => ({ id, teamId: "team" }),
  loadSprint: async (id: string) => ({ id, teamId: "team" }),
};

test("objective/sprint notifications open filtered team tasks using the resolved team", async () => {
  assert.deepEqual(
    await resolveNotificationDestination(
      { id: "n", entityId: "o", entityType: "objective" },
      context,
    ),
    {
      type: "teamStories",
      href: {
        pathname: "/teams/[teamId]",
        params: { objectiveId: "o", teamId: "team" },
      },
    },
  );
  assert.deepEqual(
    await resolveNotificationDestination(
      { id: "n", entityId: "s", entityType: "sprint" },
      context,
    ),
    {
      type: "teamStories",
      href: {
        pathname: "/teams/[teamId]",
        params: { sprintId: "s", teamId: "team" },
      },
    },
  );
  await assert.rejects(
    () =>
      resolveNotificationDestination(
        { id: "n", entityId: "o", entityType: "objective" },
        {
          ...context,
          loadObjective: async () => null,
        },
      ),
    /no longer available/,
  );
});

test("unavailable or mismatched entities never navigate to an unrelated team", async () => {
  for (const entityType of ["objective", "sprint"] as const) {
    for (const entity of [
      null,
      { id: "wrong", teamId: "team" },
      { id: "id", teamId: "" },
    ]) {
      await assert.rejects(
        () =>
          resolveNotificationDestination(
            { id: "n", entityId: "id", entityType },
            {
              ...context,
              loadObjective: async () => entity,
              loadSprint: async () => entity,
            },
          ),
        /no longer available/,
      );
    }
  }
});

test("lookup failures remain errors and task notifications keep their direct destination", async () => {
  const offline = async () => {
    throw new Error("Offline");
  };
  await assert.rejects(
    () =>
      resolveNotificationDestination(
        { id: "n", entityId: "id", entityType: "objective" },
        { ...context, loadObjective: offline },
      ),
    /Offline/,
  );
  assert.deepEqual(
    await resolveNotificationDestination(
      { id: "n", entityId: "story", entityType: "story" },
      { ...context, loadObjective: offline, loadSprint: offline },
    ),
    { type: "story", storyId: "story" },
  );
});

test("unsupported native entities open the configured workspace website", () => {
  const notification = {
    id: "n",
    entityId: "strategy",
    entityType: "strategy" as const,
  };
  assert.equal(
    notificationWebURL("https://cloud.fortyone.app", "acme", notification),
    "https://acme.fortyone.app/strategy",
  );
  assert.equal(
    notificationWebURL("https://projects.example.com", "acme", {
      ...notification,
      entityType: "key_result",
    }),
    "https://projects.example.com/acme/notifications",
  );
  assert.throws(
    () =>
      notificationWebURL(
        "https://cloud.fortyone.app",
        "acme.evil.com",
        notification,
      ),
    /invalid/,
  );
});
