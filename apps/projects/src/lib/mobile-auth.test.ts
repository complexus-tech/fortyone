import { getOnboardingWorkspaceUrl } from "@/modules/onboarding/routing";
import { getOnboardingStartUrl } from "@/modules/onboarding/start";
import { getWelcomeDestinations } from "@/modules/onboarding/welcome/destinations";
import { withCallbackUrl } from "@/utils/callback-url";
import {
  getMobileAuthPath,
  getMobileRedirectURL,
  isMobileAuthPath,
  parseMobileAuthRequest,
} from "./mobile-auth";

const transaction = { state: "s".repeat(43), codeChallenge: "c".repeat(43) };

describe("mobile browser handoff", () => {
  it("requires one high entropy state and S256 challenge", () => {
    expect(
      parseMobileAuthRequest({
        state: transaction.state,
        code_challenge: transaction.codeChallenge,
      }),
    ).toEqual(transaction);
    expect(
      parseMobileAuthRequest({
        state: [transaction.state],
        code_challenge: transaction.codeChallenge,
      }),
    ).toBeNull();
    expect(
      parseMobileAuthRequest({
        state: "short",
        code_challenge: transaction.codeChallenge,
      }),
    ).toBeNull();
  });

  it("only recognizes a complete relative first-party callback", () => {
    const path = getMobileAuthPath(transaction);
    expect(isMobileAuthPath(path)).toBe(true);
    expect(isMobileAuthPath(`${path}&state=${transaction.state}`)).toBe(false);
    expect(isMobileAuthPath(`${path}#fragment`)).toBe(false);
    expect(isMobileAuthPath(`https://attacker.example${path}`)).toBe(false);
    expect(isMobileAuthPath("/auth/mobile?state=short")).toBe(false);
  });

  it("returns only a code and state to the fixed native callback", () => {
    const url = new URL(
      getMobileRedirectURL("k".repeat(43), transaction.state),
    );
    expect(`${url.protocol}//${url.host}`).toBe("fortyone://login");
    expect([...url.searchParams.keys()]).toEqual(["code", "state"]);
    expect(() =>
      getMobileRedirectURL("unsafe&token=value", transaction.state),
    ).toThrow();
  });

  it("preserves mobile context through sign-in and onboarding completion", () => {
    const path = getMobileAuthPath(transaction);
    expect(
      new URL(
        withCallbackUrl("/auth-callback?mobileApp=true", path),
        "https://cloud.fortyone.app",
      ).searchParams.get("callbackUrl"),
    ).toBe(path);
    expect(getOnboardingWorkspaceUrl("example", path)).toBe(path);
    expect(getOnboardingStartUrl("example", "task", path)).toBe(path);
    expect(
      getWelcomeDestinations(
        [{ id: "workspace", slug: "example", userRole: "member" }],
        "workspace",
        path,
      ).redirectUrl,
    ).toBe(path);
  });
});
