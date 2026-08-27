/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Team } from "@/modules/teams/types";
import type { Workspace } from "@/types";
import { getUserContext } from "./user-context";

const workspace: Workspace = {
  id: "workspace-1",
  name: "Acme",
  slug: "acme",
  color: "#000000",
  avatarUrl: null,
  userRole: "admin",
  trialEndsOn: null,
  deletedAt: null,
  isActive: true,
  createdAt: "2026-01-01T00:00:00.000Z",
  updatedAt: "2026-01-01T00:00:00.000Z",
};

const joinedTeams: Team[] = [
  {
    id: "team-1",
    name: "Product",
    code: "PROD",
    color: "#000000",
    isPrivate: false,
    workspaceId: workspace.id,
    createdAt: "2026-01-01T00:00:00.000Z",
    updatedAt: "2026-01-01T00:00:00.000Z",
    memberCount: 1,
    sprintsEnabled: true,
  },
];

const contextInput = {
  currentPath: "/stories",
  currentTheme: "system",
  resolvedTheme: "dark",
  subscription: {
    tier: "pro",
    billingInterval: "monthly",
    billingEndsAt: "2026-08-01",
    status: "active",
  },
  joinedTeams,
  memories: [],
  terminology: {
    stories: "Stories",
    sprints: "Sprints",
    objectives: "Objectives",
    keyResults: "Key Results",
  },
  workspace,
  totalMessages: {
    current: 2,
    limit: 100,
  },
};

describe("getUserContext", () => {
  it("returns no context when the request is unauthenticated", () => {
    expect(getUserContext(contextInput)).toBe("");
  });

  it("uses the authenticated identity without performing another lookup", () => {
    const context = getUserContext({
      ...contextInput,
      user: {
        id: "user-1",
        name: "Maya",
      },
      username: "maya",
      storyCreationDefaults: {
        autoSchedulingAvailable: true,
        bulkStories: {
          autoSchedulingEnabled: false,
          estimatedDurationMinutes: null,
        },
        singleStory: {
          autoSchedulingEnabled: true,
          estimatedDurationMinutes: 60,
        },
      },
    });

    expect(context).toContain("User: Maya (@maya) [user-1]");
    expect(context).toContain("resolve to Maya [user-1].");
    expect(context).toContain("Product (PROD) [team-1]");
    expect(context).toContain("Joined teams:");
    expect(context).toContain(
      'If the user says "my team" without naming one, use Product [team-1].',
    );
    expect(context).not.toContain("Growth");
    expect(context).toContain(
      "Never treat a discoverable public team as one of the user's teams.",
    );
    expect(context).toContain(
      "Account story-creation defaults: single-story suggestions: time needed=60 minutes, calendar scheduling=on; multiple-story default: no shared time estimate, calendar scheduling=off",
    );
    expect(context).toContain(
      "Single-story defaults are suggestions, not consent.",
    );
    expect(context).toContain(
      "Never apply or mention them as batch-wide defaults.",
    );
  });

  it("does not substitute public teams when the user has no memberships", () => {
    const context = getUserContext({
      ...contextInput,
      joinedTeams: [],
      user: {
        id: "user-1",
        name: "Maya",
      },
    });

    expect(context).toContain("Joined teams:\n- None");
    expect(context).toContain(
      'The user has not joined a team. Say that plainly when they ask about "my team"; do not offer public teams as substitutes.',
    );
  });

  it("requires a membership lookup when joined teams are unavailable", () => {
    const context = getUserContext({
      ...contextInput,
      joinedTeams: null,
      user: {
        id: "user-1",
        name: "Maya",
      },
    });

    expect(context).toContain("Joined teams:\n- Unavailable");
    expect(context).toContain(
      'Joined-team membership could not be loaded. Call listTeams before answering a request about "my team"',
    );
    expect(context).toContain(
      "do not infer membership from accessible or public teams",
    );
  });

  it("does not render an undefined username", () => {
    const context = getUserContext({
      ...contextInput,
      user: {
        id: "user-1",
        name: "Maya",
      },
    });

    expect(context).toContain("User: Maya [user-1]");
    expect(context).not.toContain("@undefined");
    expect(context).toContain(
      "Account story-creation defaults: single-story suggestions unavailable; multiple-story default: no shared time estimate, calendar scheduling=off",
    );
  });

  it("bounds dynamic memories and omits billing metadata from model context", () => {
    const context = getUserContext({
      ...contextInput,
      memories: Array.from({ length: 20 }, (_, index) => ({
        id: `memory-${index}`,
        content: `Memory ${index} ${"x".repeat(1000)}`,
        workspaceId: workspace.id,
        userId: "user-1",
        createdAt: "2026-01-01T00:00:00.000Z",
        updatedAt: "2026-01-01T00:00:00.000Z",
      })),
      user: {
        id: "user-1",
        name: "Maya",
      },
    });

    expect(context).toContain("+8 more memories omitted");
    expect(context).not.toContain("memory-12");
    expect(context).not.toContain("Billing interval");
    expect(context).not.toContain("Message usage");
  });
});
