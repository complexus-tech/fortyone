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

const teams: Team[] = [
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
  teams,
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
    });

    expect(context).toContain("User: Maya (@maya) [user-1]");
    expect(context).toContain("resolve to Maya [user-1].");
    expect(context).toContain("Product (PROD) [team-1]");
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
  });
});
