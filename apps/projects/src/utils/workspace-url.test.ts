/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { Workspace } from "@/types/workspace";
import { getMobileAuthPath } from "@/lib/mobile-auth";
import type * as WorkspaceRouting from "./workspace-url";

describe.each([
  ["fortyone.app", "https://acme.fortyone.app"],
  ["localhost", "/acme"],
])("workspace defaults on %s", (domain, base) => {
  const loadRouting = () => {
    const original = process.env.NEXT_PUBLIC_DOMAIN;
    let routing: typeof WorkspaceRouting;
    try {
      process.env.NEXT_PUBLIC_DOMAIN = domain;
      jest.isolateModules(() => {
        routing =
          jest.requireActual<typeof WorkspaceRouting>("./workspace-url");
      });
    } finally {
      if (original === undefined) delete process.env.NEXT_PUBLIC_DOMAIN;
      else process.env.NEXT_PUBLIC_DOMAIN = original;
    }
    return routing!;
  };
  const workspaces = [
    { id: "first", slug: "first" },
    { id: "active", slug: "acme" },
  ] as Workspace[];

  it("uses Maya for workspace links and the last-used workspace", () => {
    const { buildWorkspaceUrl, getRedirectUrl } = loadRouting();
    expect(buildWorkspaceUrl("acme")).toBe(`${base}/maya`);
    expect(getRedirectUrl(workspaces, [], "active")).toBe(`${base}/maya`);
    expect(getRedirectUrl([workspaces[1]])).toBe(`${base}/maya`);
  });

  it("preserves explicit destinations and callbacks", () => {
    const { buildWorkspaceUrl, getRedirectUrl } = loadRouting();
    expect(buildWorkspaceUrl("acme", "/my-work")).toBe(`${base}/my-work`);
    expect(
      getRedirectUrl(workspaces, [], "active", "/my-work?view=board"),
    ).toBe("/my-work?view=board");
  });

  it.each([
    "/acme/my-work?view=board",
    "/my-work?view=board",
    "https://acme.fortyone.app/my-work?view=board",
    "https://acme.fortyone.app/",
    "/auth-callback",
    "/onboarding/account",
  ])(
    "starts workspace creation instead of returning to %s without membership",
    (callback) => {
      const { getRedirectUrl } = loadRouting();
      expect(getRedirectUrl([], [], "deleted-workspace", callback)).toBe(
        "/onboarding/create",
      );
    },
  );

  it.each([
    "/account",
    "/profile",
    "/portal/community/feedback",
    "https://community.fortyone.app/feedback?newFeedback=true",
    "https://community.fortyone.app/account",
    "/onboarding/join?token=invitation",
    "/onboarding/create?callbackUrl=%2Fsettings%2Fworkspace%2Ffeedback",
    "https://api.fortyone.app/oauth/authorize?client_id=external-app",
    "/github/callback?code=code&state=state",
    "/auth/account-deletion",
    getMobileAuthPath({ state: "s".repeat(43), codeChallenge: "c".repeat(43) }),
  ])("preserves workspace-independent continuation to %s", (callback) => {
    const { getRedirectUrl } = loadRouting();
    expect(getRedirectUrl([], [], undefined, callback)).toBe(callback);
  });

  it("offers an available invitation when the former workspace callback is stale", () => {
    const { getRedirectUrl } = loadRouting();
    expect(
      getRedirectUrl(
        [],
        [{}, { token: "invite&token" }],
        "deleted-workspace",
        "https://deleted.fortyone.app/maya",
      ),
    ).toBe("/onboarding/join?token=invite%26token");
  });
});
