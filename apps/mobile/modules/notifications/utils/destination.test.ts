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

test("notification routing resolves objective/sprint teams and never treats them as tasks", async () => {
  assert.deepEqual(
    await resolveNotificationDestination(
      { id: "n", entityId: "o", entityType: "objective" },
      context,
    ),
    {
      type: "objective",
      objectiveId: "o",
      teamId: "team",
    },
  );
  assert.deepEqual(
    await resolveNotificationDestination(
      { id: "n", entityId: "s", entityType: "sprint" },
      context,
    ),
    {
      type: "sprint",
      sprintId: "s",
      teamId: "team",
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
