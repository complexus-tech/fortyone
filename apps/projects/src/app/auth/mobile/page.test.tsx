import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import MobileAuthPage from "./page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/get-workspaces", () => ({ getWorkspaces: jest.fn() }));
jest.mock("@/modules/invitations/public/server", () => ({
  getMyInvitationsForCurrentRequest: jest.fn(),
}));
jest.mock("@/components/layouts/onboarding-layout", () => ({
  OnboardingLayout: () => null,
}));
jest.mock("./client", () => ({ MobileAuthorization: () => null }));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

const params = { state: "s".repeat(43), code_challenge: "c".repeat(43) };
const callback = `/auth/mobile?state=${params.state}&code_challenge=${params.code_challenge}`;

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(auth).mockReset();
  jest.mocked(getWorkspaces).mockResolvedValue([]);
  jest.mocked(getMyInvitationsForCurrentRequest).mockResolvedValue([]);
});

it("keeps transaction context while asking an unauthenticated user to sign in", async () => {
  jest.mocked(auth).mockResolvedValue(null);
  await expect(
    MobileAuthPage({ searchParams: Promise.resolve(params) }),
  ).rejects.toThrow(
    `redirect:/?mobileApp=true&callbackUrl=${encodeURIComponent(callback)}`,
  );
});

it("routes a new account through workspace onboarding with its return context", async () => {
  jest
    .mocked(auth)
    .mockResolvedValue({
      user: { id: "user", email: "member@example.com" },
    } as Awaited<ReturnType<typeof auth>>);
  await expect(
    MobileAuthPage({ searchParams: Promise.resolve(params) }),
  ).rejects.toThrow(
    `redirect:/onboarding/create?callbackUrl=${encodeURIComponent(callback)}`,
  );
});

it("rejects malformed transactions without checking an account or issuing a code", async () => {
  await MobileAuthPage({ searchParams: Promise.resolve({ state: "invalid" }) });
  expect(auth).not.toHaveBeenCalled();
  expect(getWorkspaces).not.toHaveBeenCalled();
});
