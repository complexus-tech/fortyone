import { getOnboardingWorkspaceUrl } from "@/modules/onboarding/routing";
import { getOnboardingStartUrl } from "@/modules/onboarding/start";
import { getWelcomeDestinations } from "@/modules/onboarding/welcome/destinations";
import {
  getAuthCallbackPath,
  getLoginUrl,
  getSafeCallbackUrl,
  withCallbackUrl,
} from "@/utils/callback-url";
import { getRedirectUrl } from "@/utils/workspace-url";
import {
  DESKTOP_DEV_REDIRECT_URI,
  DESKTOP_REDIRECT_URI,
  getDesktopAuthPath,
  getDesktopRedirectURL,
  isDesktopAuthPath,
  parseDesktopAuthRequest,
} from "./desktop-auth";

const transaction = { state: "s".repeat(43), codeChallenge: "c".repeat(43) };

afterEach(() => {
  jest.restoreAllMocks();
});

describe("desktop browser authorization", () => {
  it("accepts one well-formed state and S256 challenge and rejects repeated parameters", () => {
    expect(
      parseDesktopAuthRequest({
        state: transaction.state,
        code_challenge: transaction.codeChallenge,
      }),
    ).toEqual(transaction);
    for (const state of ["short", [transaction.state], undefined]) {
      expect(
        parseDesktopAuthRequest({
          state,
          code_challenge: transaction.codeChallenge,
        }),
      ).toBeNull();
    }
    expect(
      parseDesktopAuthRequest({
        state: transaction.state,
        code_challenge: [transaction.codeChallenge],
      }),
    ).toBeNull();
  });

  it("only recognizes an exact relative desktop transaction", () => {
    const path = getDesktopAuthPath(transaction);
    expect(isDesktopAuthPath(path)).toBe(true);
    for (const invalid of [
      `https://attacker.example${path}`,
      `//attacker.example${path}`,
      `${path}#fragment`,
      `${path}&state=${transaction.state}`,
      `${path}&redirectUri=https://attacker.example`,
      "/auth/desktop?state=short",
      `/auth/desktop/extra?${path.split("?")[1]}`,
    ])
      expect(isDesktopAuthPath(invalid)).toBe(false);
  });

  it("returns only a code and state to the fixed desktop scheme", () => {
    const result = new URL(
      getDesktopRedirectURL("k".repeat(43), transaction.state),
    );
    expect(`${result.protocol}//${result.host}`).toBe(
      "fortyone-desktop://login",
    );
    expect([...result.searchParams.keys()]).toEqual(["code", "state"]);
    expect(() =>
      getDesktopRedirectURL("unsafe&token=value", transaction.state),
    ).toThrow();
    expect(getSafeCallbackUrl(result.toString())).toBeUndefined();
  });

  it("preserves browser OAuth choices through login and onboarding continuations", () => {
    const path = getDesktopAuthPath(transaction);
    for (const callback of [
      path,
      withCallbackUrl("/onboarding/join?token=invite", path),
    ]) {
      for (const destination of [
        getLoginUrl(callback),
        getAuthCallbackPath(callback),
      ]) {
        const url = new URL(destination, "https://cloud.fortyone.app");
        expect(url.searchParams.get("mobileApp")).toBeNull();
        expect(url.searchParams.get("callbackUrl")).toBe(callback);
      }
    }
    expect(getRedirectUrl([], [], undefined, path)).toBe(path);
    expect(getOnboardingWorkspaceUrl("acme", path)).toBe(path);
    expect(getOnboardingStartUrl("acme", "task", path)).toBe(path);
    expect(
      getWelcomeDestinations(
        [{ id: "one", slug: "acme", userRole: "member" }],
        "one",
        path,
      ),
    ).toEqual({ redirectUrl: path });
  });

  it("preserves the exact development callback through browser continuations", () => {
    jest.replaceProperty(process, "env", {
      ...process.env,
      NODE_ENV: "development",
    });
    const request = parseDesktopAuthRequest({
      state: transaction.state,
      code_challenge: transaction.codeChallenge,
      redirect_uri: DESKTOP_DEV_REDIRECT_URI,
    });
    expect(request).toEqual({
      ...transaction,
      redirectUri: DESKTOP_DEV_REDIRECT_URI,
    });
    const path = getDesktopAuthPath(request!);
    expect(isDesktopAuthPath(path)).toBe(true);
    expect(getOnboardingWorkspaceUrl("acme", path)).toBe(path);
    expect(
      new URL(getLoginUrl(path), "http://localhost:3000").searchParams.get(
        "callbackUrl",
      ),
    ).toBe(path);
    expect(
      new URL(
        getDesktopRedirectURL(
          "k".repeat(43),
          transaction.state,
          DESKTOP_DEV_REDIRECT_URI,
        ),
      ).protocol,
    ).toBe("fortyone-desktop-dev:");
    for (const redirectUri of [
      "",
      [DESKTOP_DEV_REDIRECT_URI],
      "https://attacker.example",
      `${DESKTOP_DEV_REDIRECT_URI}/extra`,
      `${DESKTOP_DEV_REDIRECT_URI}?token=secret`,
    ]) {
      expect(
        parseDesktopAuthRequest({
          state: transaction.state,
          code_challenge: transaction.codeChallenge,
          redirect_uri: redirectUri,
        }),
      ).toBeNull();
    }
    expect(
      isDesktopAuthPath(
        `${path}&redirect_uri=${encodeURIComponent(DESKTOP_REDIRECT_URI)}`,
      ),
    ).toBe(false);
  });

  it("never enables the development callback in a production browser build", () => {
    jest.replaceProperty(process, "env", {
      ...process.env,
      NODE_ENV: "production",
    });
    const params = {
      state: transaction.state,
      code_challenge: transaction.codeChallenge,
    };
    expect(
      parseDesktopAuthRequest({
        ...params,
        redirect_uri: DESKTOP_DEV_REDIRECT_URI,
      }),
    ).toBeNull();
    expect(
      isDesktopAuthPath(
        getDesktopAuthPath({
          ...transaction,
          redirectUri: DESKTOP_DEV_REDIRECT_URI,
        }),
      ),
    ).toBe(false);
    expect(() =>
      getDesktopRedirectURL(
        "k".repeat(43),
        transaction.state,
        DESKTOP_DEV_REDIRECT_URI,
      ),
    ).toThrow();
    expect(
      parseDesktopAuthRequest({
        ...params,
        redirect_uri: DESKTOP_REDIRECT_URI,
      }),
    ).toEqual({ ...transaction, redirectUri: DESKTOP_REDIRECT_URI });
  });
});
