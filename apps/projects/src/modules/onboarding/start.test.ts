/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { getOnboardingStartUrl } from "./start";

jest.mock("@/utils", () => ({
  buildWorkspaceUrl: (slug: string, path: string) => `/${slug}${path}`,
}));
jest.mock("@/utils/workspace-url", () => ({
  buildWorkspaceUrl: (slug: string, path: string) =>
    `https://${slug}.fortyone.app${path}`,
}));

describe("onboarding first action destinations", () => {
  it.each([
    ["task", "/my-work?onboarding=task"],
    ["import", "/settings/workspace/imports?from=onboarding"],
    ["examples", "/maya"],
    ["empty", "/maya"],
  ] as const)("uses the workspace URL builder for %s", (start, path) => {
    expect(getOnboardingStartUrl("acme", start)).toBe(
      `https://acme.fortyone.app${path}`,
    );
  });

  it("keeps an explicit workspace callback ahead of the selected start", () => {
    expect(
      getOnboardingStartUrl(
        "acme",
        "task",
        "/settings/account?tab=profile#name",
      ),
    ).toBe("https://acme.fortyone.app/settings/account?tab=profile#name");
  });

  it.each(["//example.com", "https://example.com"])(
    "falls back to the selected start for an unsupported callback: %s",
    (callback) => {
      expect(getOnboardingStartUrl("acme", "task", callback)).toBe(
        "https://acme.fortyone.app/my-work?onboarding=task",
      );
    },
  );
});
