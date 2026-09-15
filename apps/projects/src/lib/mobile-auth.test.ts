import { getOnboardingWorkspaceUrl } from "@/modules/onboarding/routing";
import { getOnboardingStartUrl } from "@/modules/onboarding/start";
import { getWelcomeDestinations } from "@/modules/onboarding/welcome/destinations";
import { withCallbackUrl } from "@/utils/callback-url";
import {
  getMobileAuthPath,
  getMobileRedirectURL,
  isMobileAuthPath,
  isMobileAuthFlow,
  parseMobileAuthRequest,
  MOBILE_ACCOUNT_DELETION_PATH,
} from "./mobile-auth";

const transaction = { state: "s".repeat(43), codeChallenge: "c".repeat(43) };

describe("mobile browser handoff", () => {
  it("keeps deletion and its exact workspace resolution routes email-only", () => {
    expect(isMobileAuthFlow(MOBILE_ACCOUNT_DELETION_PATH)).toBe(true);
    expect(
      isMobileAuthFlow(
        withCallbackUrl("/auth-callback", MOBILE_ACCOUNT_DELETION_PATH),
      ),
    ).toBe(true);
    expect(
      isMobileAuthFlow(
        withCallbackUrl("/acme/settings", MOBILE_ACCOUNT_DELETION_PATH),
      ),
    ).toBe(true);
    expect(
      isMobileAuthFlow(
        withCallbackUrl(
          "/acme/settings/workspace/members",
          MOBILE_ACCOUNT_DELETION_PATH,
        ),
      ),
    ).toBe(true);
    expect(
      isMobileAuthFlow(
        withCallbackUrl(
          "/acme/settings/workspace/billing",
          MOBILE_ACCOUNT_DELETION_PATH,
        ),
      ),
    ).toBe(false);
    expect(isMobileAuthFlow("/auth/account-deletion-unknown")).toBe(false);
    expect(
      isMobileAuthFlow("https://attacker.example/auth/account-deletion"),
    ).toBe(false);
  });
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

  it("recognizes mobile handoffs nested in an onboarding login return", () => {
    const path = getMobileAuthPath(transaction);
    expect(isMobileAuthFlow(path)).toBe(true);
    expect(
      isMobileAuthFlow(withCallbackUrl("/onboarding/join?token=invite", path)),
    ).toBe(true);
    expect(
      isMobileAuthFlow(
        withCallbackUrl(
          "/auth-callback",
          withCallbackUrl("/onboarding/account", path),
        ),
      ),
    ).toBe(true);
  });

  it("does not infer mobile mode from unrelated or unsafe return URLs", () => {
    const path = getMobileAuthPath(transaction);
    for (const callback of [
      withCallbackUrl("/settings/workspace/billing", path),
      `https://example.com${path}`,
      `//example.com${path}`,
      "/onboarding/join?callbackUrl=/auth/mobile?state=invalid",
      `${withCallbackUrl("/onboarding/join", path)}&callbackUrl=/other`,
      `${withCallbackUrl("/onboarding/join", path)}#fragment`,
    ]) {
      expect(isMobileAuthFlow(callback)).toBe(false);
    }
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
