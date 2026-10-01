/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { Session } from "@/auth";
import type { Workspace } from "@/types";
import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getWorkspaceSSOStatus } from "@/modules/auth/workspace-sso-status";
import WorkspaceSSOPage from "./page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspaces", () => ({
  getWorkspaces: jest.fn(),
}));
jest.mock("@/modules/auth/workspace-sso-status", () => ({
  getWorkspaceSSOStatus: jest.fn(),
}));
jest.mock("@/components/layouts/onboarding-layout", () => ({
  OnboardingLayout: () => null,
}));
jest.mock("@/modules/auth/workspace-sso-recovery", () => ({
  WorkspaceSSORecovery: () => null,
}));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

const session: Session = {
  user: {
    id: "member",
    email: "member@acme.example",
    name: "Member",
    fullName: "Member",
    username: "member",
    image: null,
    isInternal: false,
    lastUsedWorkspaceId: "acme-workspace",
  },
};
const workspace: Workspace = {
  id: "acme-workspace",
  name: "Acme",
  slug: "acme",
  color: "blue",
  avatarUrl: null,
  userRole: "member",
  trialEndsOn: null,
  deletedAt: null,
  isActive: true,
  createdAt: "2026-10-01T00:00:00Z",
  updatedAt: "2026-10-01T00:00:00Z",
};
const params = Promise.resolve({ workspaceSlug: "acme" });

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(auth).mockResolvedValue(session);
  jest.mocked(getWorkspaces).mockResolvedValue([workspace]);
  jest
    .mocked(getWorkspaceSSOStatus)
    .mockResolvedValue({ enabled: true, requireSSO: true });
});

describe("workspace SSO recovery", () => {
  it("holds the authenticated account and renders the public SSO continuation outside the blocked workspace layout", async () => {
    const page = await WorkspaceSSOPage({ params });
    expect(page.props.children.props).toEqual({
      email: session.user.email,
      required: true,
      workspaceName: "Acme",
      workspaceSlug: "acme",
    });
    expect(auth).toHaveBeenCalledTimes(1);
  });

  it("returns an unlinked member through ordinary account sign-in before continuing SSO", async () => {
    jest.mocked(auth).mockResolvedValue(null);
    await expect(WorkspaceSSOPage({ params })).rejects.toThrow(
      "callbackUrl=%2Fauth%2Fsso%2Facme",
    );
    expect(getWorkspaces).not.toHaveBeenCalled();
    expect(getWorkspaceSSOStatus).not.toHaveBeenCalled();
  });

  it("rejects a current account without membership even if public SSO is enabled", async () => {
    jest.mocked(getWorkspaces).mockResolvedValue([]);
    await expect(WorkspaceSSOPage({ params })).rejects.toThrow(
      "redirect:/unauthorized",
    );
  });
});
