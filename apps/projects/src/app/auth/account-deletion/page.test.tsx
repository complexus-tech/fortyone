import { auth } from "@/auth";
import { MOBILE_ACCOUNT_DELETION_PATH } from "@/lib/mobile-auth";
import AccountDeletionPage from "./page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/modules/settings/account/delete", () => ({
  DeleteAccountSettings: () => null,
}));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

beforeEach(() => {
  jest.clearAllMocks();
});

it("uses email-only login with the deletion return path for missing or expired sessions", async () => {
  jest.mocked(auth).mockResolvedValue(null);
  await expect(AccountDeletionPage()).rejects.toThrow(
    `mobileApp=true&callbackUrl=${encodeURIComponent(MOBILE_ACCOUNT_DELETION_PATH)}`,
  );
});

it("renders the verified browser identity without requiring any workspace", async () => {
  jest
    .mocked(auth)
    .mockResolvedValue({
      user: { id: "current-user", email: "current@example.com" },
    } as Awaited<ReturnType<typeof auth>>);
  const page = await AccountDeletionPage();
  expect(page.props.children.key).toBe("current-user");
  expect(page.props.children.props).toEqual({
    account: { id: "current-user", email: "current@example.com" },
    mobileAuth: true,
  });
});
