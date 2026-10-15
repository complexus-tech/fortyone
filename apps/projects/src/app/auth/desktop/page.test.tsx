import { auth } from "@/auth";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import { getMyInvitationsForCurrentRequest } from "@/modules/invitations/public/server";
import DesktopAuthPage from "./page";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/get-workspaces", () => ({ getWorkspaces: jest.fn() }));
jest.mock("@/modules/invitations/public/server", () => ({
  getMyInvitationsForCurrentRequest: jest.fn(),
}));
jest.mock("@/components/layouts/onboarding-layout", () => ({
  OnboardingLayout: () => null,
}));
jest.mock("./client", () => ({ DesktopAuthorization: () => null }));
jest.mock("next/navigation", () => ({
  redirect: (path: string) => {
    throw new Error(`redirect:${path}`);
  },
}));

const params = { state: "s".repeat(43), code_challenge: "c".repeat(43) };
const callback = `/auth/desktop?state=${params.state}&code_challenge=${params.code_challenge}`;

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(auth).mockReset();
  jest.mocked(getWorkspaces).mockResolvedValue([]);
  jest.mocked(getMyInvitationsForCurrentRequest).mockResolvedValue([]);
});

afterEach(() => {
  jest.restoreAllMocks();
});

it("keeps the validated development callback through browser sign-in", async () => {
  jest.replaceProperty(process, "env", {
    ...process.env,
    NODE_ENV: "development",
  });
  jest.mocked(auth).mockResolvedValue(null);
  const redirectUri = "fortyone-desktop-dev://login";
  const target = `${callback}&redirect_uri=${encodeURIComponent(redirectUri)}`;
  await expect(
    DesktopAuthPage({
      searchParams: Promise.resolve({ ...params, redirect_uri: redirectUri }),
    }),
  ).rejects.toThrow(`redirect:/?callbackUrl=${encodeURIComponent(target)}`);
});

it("preserves the desktop transaction on regular browser login without mobile restrictions", async () => {
  jest.mocked(auth).mockResolvedValue(null);
  await expect(
    DesktopAuthPage({ searchParams: Promise.resolve(params) }),
  ).rejects.toThrow(`redirect:/?callbackUrl=${encodeURIComponent(callback)}`);
});

it("keeps the transaction through new-workspace onboarding", async () => {
  jest.mocked(auth).mockResolvedValue({
    user: { id: "user", email: "member@example.com" },
  } as Awaited<ReturnType<typeof auth>>);
  await expect(
    DesktopAuthPage({ searchParams: Promise.resolve(params) }),
  ).rejects.toThrow(
    `redirect:/onboarding/create?callbackUrl=${encodeURIComponent(callback)}`,
  );
});

it("preserves invitation onboarding when a desktop user has no workspace", async () => {
  jest.mocked(auth).mockResolvedValue({
    user: { id: "user", email: "member@example.com" },
  } as Awaited<ReturnType<typeof auth>>);
  jest
    .mocked(getMyInvitationsForCurrentRequest)
    .mockResolvedValue([{ token: "invite" }] as Awaited<
      ReturnType<typeof getMyInvitationsForCurrentRequest>
    >);
  await expect(
    DesktopAuthPage({ searchParams: Promise.resolve(params) }),
  ).rejects.toThrow(
    `redirect:/onboarding/join?token=invite&callbackUrl=${encodeURIComponent(callback)}`,
  );
});

it("rejects an invalid transaction before account lookup or code issuance", async () => {
  await DesktopAuthPage({
    searchParams: Promise.resolve({
      state: [params.state],
      code_challenge: params.code_challenge,
    }),
  });
  expect(auth).not.toHaveBeenCalled();
  expect(getWorkspaces).not.toHaveBeenCalled();
});

it("requires explicit confirmation of the signed-in browser identity", async () => {
  jest.mocked(auth).mockResolvedValue({
    user: { id: "user", email: "member@example.com" },
  } as Awaited<ReturnType<typeof auth>>);
  jest
    .mocked(getWorkspaces)
    .mockResolvedValue([{ id: "workspace", slug: "acme" }] as Awaited<
      ReturnType<typeof getWorkspaces>
    >);
  const page = await DesktopAuthPage({ searchParams: Promise.resolve(params) });
  expect(page.props.children.props).toEqual({
    email: "member@example.com",
    transaction: { state: params.state, codeChallenge: params.code_challenge },
  });
});
