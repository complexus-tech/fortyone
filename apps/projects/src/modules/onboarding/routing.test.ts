/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import {
  getOnboardingCallbackPath,
  getOnboardingWorkspaceUrl,
  withOnboardingCallbackUrl,
} from "./routing";

jest.mock("@/utils", () => ({
  buildWorkspaceUrl: (workspaceSlug: string, path = "/my-work") =>
    `/${workspaceSlug}${path}`,
}));

describe("onboarding callback routing", () => {
  it("carries a safe workspace path to the next onboarding step", () => {
    expect(
      withOnboardingCallbackUrl(
        "/onboarding/account",
        "/settings/workspace/feedback",
      ),
    ).toBe(
      "/onboarding/account?callbackUrl=%2Fsettings%2Fworkspace%2Ffeedback",
    );
  });

  it("rejects callbacks that cannot be applied within a workspace", () => {
    expect(getOnboardingCallbackPath("https://example.com/feedback")).toBe(
      undefined,
    );
    expect(
      withOnboardingCallbackUrl("/onboarding/invite", "//example.com"),
    ).toBe("/onboarding/invite");
  });

  it("applies the callback path to the newly created workspace", () => {
    expect(
      getOnboardingWorkspaceUrl("art-circles", "/settings/workspace/feedback"),
    ).toMatch(
      /art-circles(?:\.fortyone\.app)?\/settings\/workspace\/feedback$/,
    );
  });

  it("falls back to the workspace home when no callback is present", () => {
    expect(getOnboardingWorkspaceUrl("art-circles")).toMatch(
      /art-circles(?:\.fortyone\.app)?\/my-work$/,
    );
  });
});
