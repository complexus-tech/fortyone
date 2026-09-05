/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { Workspace } from "@/types/workspace";
import { getWelcomeDestinations } from "./destinations";

jest.mock("@/utils", () => ({
  buildWorkspaceUrl: (workspaceSlug: string, path = "/maya") =>
    `/${workspaceSlug}${path}`,
}));
jest.mock("@/utils/workspace-url", () => ({
  buildWorkspaceUrl: (workspaceSlug: string, path = "/maya") =>
    `/${workspaceSlug}${path}`,
}));

const workspace = (id: string, slug: string, userRole: Workspace["userRole"]) =>
  ({ id, slug, userRole }) satisfies Pick<
    Workspace,
    "id" | "slug" | "userRole"
  >;

describe("welcome destinations", () => {
  it("offers the fixed onboarding import route for the active admin workspace", () => {
    const destinations = getWelcomeDestinations(
      [workspace("one", "first", "admin"), workspace("two", "active", "admin")],
      "two",
      "/settings/account",
    );

    expect(destinations).toEqual({
      redirectUrl: "/active/settings/account",
      importUrl: "/active/settings/workspace/imports?from=onboarding",
      taskUrl: "/active/my-work?onboarding=task",
      calendarUrl: "/active/settings/account/calendar",
    });
  });

  it.each(["member", "guest"] as const)(
    "does not offer imports to a %s workspace",
    (userRole) => {
      expect(
        getWelcomeDestinations([workspace("workspace", "acme", userRole)]),
      ).toEqual({
        redirectUrl: "/acme/maya",
        calendarUrl: "/acme/settings/account/calendar",
        ...(userRole === "member"
          ? { taskUrl: "/acme/my-work?onboarding=task" }
          : {}),
      });
    },
  );

  it("preserves the callback while routing workspace-less users to creation", () => {
    expect(getWelcomeDestinations([], undefined, "/settings/account")).toEqual({
      redirectUrl: "/onboarding/create?callbackUrl=%2Fsettings%2Faccount",
    });
  });
});
