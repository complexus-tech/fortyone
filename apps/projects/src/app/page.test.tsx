/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { auth } from "@/auth";
import { getProfile } from "@/lib/queries/profile";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import LoginPage from "./page";
import SignupPage from "./signup/page";
import VerifyPage from "./verify/[email]/[token]/page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/get-workspaces", () => ({ getWorkspaces: jest.fn() }));
jest.mock("@/lib/queries/profile", () => ({ getProfile: jest.fn() }));
jest.mock("@/modules/invitations/public/server", () => ({
  getMyInvitationsForCurrentRequest: jest.fn(),
}));
jest.mock("@/components/layouts/onboarding-layout", () => ({
  OnboardingLayout: () => null,
}));
jest.mock("@/modules/auth", () => ({ AuthLayout: () => null }));
jest.mock("./verify/[email]/[token]/client", () => ({
  EmailVerificationCallback: () => null,
}));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(auth).mockResolvedValue({
    user: { id: "returning-user", email: "returning@example.com" },
  } as Awaited<ReturnType<typeof auth>>);
  jest.mocked(getProfile).mockResolvedValue({
    id: "returning-user",
    fullName: "Returning User",
    lastUsedWorkspaceId: "former-workspace",
  } as Awaited<ReturnType<typeof getProfile>>);
  jest.mocked(getWorkspaces).mockResolvedValue([]);
  jest.mocked(getMyInvitationsForCurrentRequest).mockResolvedValue([]);
});

describe.each([
  ["login", LoginPage],
  ["signup", SignupPage],
  ["email verification", VerifyPage],
] as const)("returning accounts at %s", (_name, Page) => {
  it.each([undefined, "https://former.fortyone.app/maya", "/former/my-work"])(
    "opens the existing workspace creation flow when no memberships remain (%s)",
    async (callbackUrl) => {
      await expect(
        Page({ searchParams: Promise.resolve({ callbackUrl }) }),
      ).rejects.toThrow("redirect:/onboarding/create");
    },
  );
});

it.each([
  "account_unavailable",
  "oauth_expired",
  "oauth_failed",
  "oauth_cancelled",
])(
  "shows %s without bouncing through an existing browser session",
  async (error) => {
    const page = await LoginPage({
      searchParams: Promise.resolve({
        callbackUrl: "/former/my-work",
        error,
      }),
    });
    expect(page.props.children.props).toMatchObject({
      page: "login",
      errorMessage: error,
      callbackUrl: "/former/my-work",
    });
    expect(getWorkspaces).not.toHaveBeenCalled();
    expect(getProfile).not.toHaveBeenCalled();
  },
);
