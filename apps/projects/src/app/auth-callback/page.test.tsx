import { auth } from "@/auth";
import { MOBILE_ACCOUNT_DELETION_PATH } from "@/lib/mobile-auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getProfile } from "@/lib/queries/profile";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import AuthCallback from "./page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/get-workspaces", () => ({
  getWorkspaces: jest.fn(),
}));
jest.mock("@/lib/queries/profile", () => ({ getProfile: jest.fn() }));
jest.mock("@/modules/invitations/public/server", () => ({
  getMyInvitationsForCurrentRequest: jest.fn(),
}));
jest.mock("./client", () => ({ ClientPage: () => null }));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

beforeEach(() => {
  jest.clearAllMocks();
});

it("returns to deletion after OTP without sending a zero-workspace account through onboarding", async () => {
  jest.mocked(auth).mockResolvedValue({
    user: { id: "current-user", email: "current@example.com" },
  } as Awaited<ReturnType<typeof auth>>);
  jest.mocked(getWorkspaces).mockResolvedValue([]);

  await expect(
    AuthCallback({
      searchParams: Promise.resolve({
        callbackUrl: MOBILE_ACCOUNT_DELETION_PATH,
      }),
    }),
  ).rejects.toThrow(`redirect:${MOBILE_ACCOUNT_DELETION_PATH}`);
  expect(getWorkspaces).not.toHaveBeenCalled();
  expect(getProfile).not.toHaveBeenCalled();
  expect(getMyInvitationsForCurrentRequest).not.toHaveBeenCalled();
});

it("preserves email-only deletion continuation when the callback session expires", async () => {
  jest.mocked(auth).mockResolvedValue(null);

  await expect(
    AuthCallback({
      searchParams: Promise.resolve({
        callbackUrl: MOBILE_ACCOUNT_DELETION_PATH,
      }),
    }),
  ).rejects.toThrow(
    `mobileApp=true&callbackUrl=${encodeURIComponent(MOBILE_ACCOUNT_DELETION_PATH)}`,
  );
  expect(getWorkspaces).not.toHaveBeenCalled();
});
