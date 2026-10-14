import { QueryClient } from "@tanstack/react-query";
import { redirect } from "next/navigation";
import { getQueryClient } from "@/app/get-query-client";
import { auth } from "@/auth";
import { workspaceKeys } from "@/constants/keys";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import { getWorkspaceAccess } from "@/modules/auth/workspace-access";
import RootLayout from "./layout";
import { fetchNonCriticalImportantQueries } from "./non-critical-important-queries";

jest.mock("next/navigation", () => ({ redirect: jest.fn() }));
jest.mock("@/app/get-query-client", () => ({ getQueryClient: jest.fn() }));
jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http/header", () => ({
  getCookieHeader: jest.fn().mockResolvedValue("session=current-user"),
}));
jest.mock("@/lib/queries/workspaces/get-workspaces", () => ({
  getWorkspaces: jest.fn(),
}));
jest.mock("@/modules/auth/workspace-access", () => ({
  getWorkspaceAccess: jest.fn(),
}));
jest.mock("./non-critical-important-queries", () => ({
  fetchNonCriticalImportantQueries: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-teams", () => ({
  getTeams: jest.fn().mockResolvedValue([]),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn().mockResolvedValue([]),
}));
jest.mock("@/modules/objectives/queries/statuses", () => ({
  getObjectiveStatuses: jest.fn().mockResolvedValue([]),
}));
jest.mock("@/lib/queries/workspaces/get-settings", () => ({
  getWorkspaceSettings: jest.fn().mockResolvedValue({}),
}));
jest.mock("@/modules/sprints/queries/get-running-sprints", () => ({
  getRunningSprints: jest.fn().mockResolvedValue([]),
}));
jest.mock("@/lib/queries/users/profile", () => ({
  getProfile: jest.fn().mockResolvedValue({}),
}));
jest.mock("@/shared/walkthrough/progress/get-progress", () => ({
  getOnboardingTourProgress: jest.fn().mockResolvedValue({}),
}));
jest.mock("@/components/ui/chat", () => ({ Chat: () => null }));
jest.mock("@/components/walkthrough/walkthrough-provider", () => ({
  WalkthroughProvider: () => null,
}));
jest.mock("@/components/walkthrough/walkthrough-tour", () => ({
  WORKSPACE_GETTING_STARTED_TOUR_KEY: "getting-started",
  WORKSPACE_GETTING_STARTED_TOUR_VERSION: 1,
}));
jest.mock("@/context/chat-context", () => ({ ChatProvider: () => null }));
jest.mock("../server-sent-events", () => ({ ServerSentEvents: () => null }));
jest.mock("./identify", () => ({ IdentifyUser: () => null }));

describe("workspace shell query reuse", () => {
  let queryClient: QueryClient;
  const workspaces = [{ id: "workspace-id", slug: "acme", name: "Acme" }];

  beforeEach(() => {
    jest.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false, gcTime: Infinity } },
    });
    jest.mocked(getQueryClient).mockReturnValue(queryClient);
    jest.mocked(auth).mockResolvedValue(null);
    jest
      .mocked(getWorkspaces)
      .mockResolvedValue(
        workspaces as Awaited<ReturnType<typeof getWorkspaces>>,
      );
    jest.mocked(getWorkspaceAccess).mockResolvedValue("allowed");
    jest.mocked(redirect).mockImplementation((url) => {
      throw new Error(`redirect:${url}`);
    });
  });

  afterEach(() => {
    queryClient.clear();
  });

  it("hydrates the authorized membership lookup without querying it again", async () => {
    await RootLayout({
      children: null,
      params: Promise.resolve({ workspaceSlug: "acme" }),
    });

    expect(getWorkspaces).toHaveBeenCalledTimes(1);
    expect(queryClient.getQueryData(workspaceKeys.lists())).toEqual(workspaces);
    expect(getWorkspaceAccess).toHaveBeenCalledWith(
      expect.objectContaining({ workspaceSlug: "acme" }),
    );
    expect(fetchNonCriticalImportantQueries).toHaveBeenCalledTimes(1);
  });

  it("does not populate or prefetch the shell before scoped access is allowed", async () => {
    jest.mocked(getWorkspaceAccess).mockResolvedValue("sso-required");

    await expect(
      RootLayout({
        children: null,
        params: Promise.resolve({ workspaceSlug: "acme" }),
      }),
    ).rejects.toThrow("redirect:/auth/sso/acme");

    expect(getQueryClient).not.toHaveBeenCalled();
    expect(fetchNonCriticalImportantQueries).not.toHaveBeenCalled();
    expect(queryClient.getQueryData(workspaceKeys.lists())).toBeUndefined();
  });
});
