import { act, renderHook } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useParams } from "next/navigation";
import { teamKeys } from "@/constants/keys";
import { useWorkspaceSettings } from "@/lib/hooks/workspace/settings";
import { useTerminology } from "./use-terminology-display";

jest.mock("next/navigation", () => ({ useParams: jest.fn() }));
jest.mock("@/lib/hooks/workspace/settings", () => ({
  useWorkspaceSettings: jest.fn(),
}));
jest.mock("./use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "first" }),
}));

let queryClient: QueryClient;
const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

describe("team work terminology", () => {
  beforeEach(() => {
    queryClient = new QueryClient();
    jest.mocked(useParams).mockReturnValue({ teamId: "sales" });
    jest.mocked(useWorkspaceSettings).mockReturnValue({
      data: {
        storyTerm: "task",
        sprintTerm: "cycle",
        objectiveTerm: "project",
        keyResultTerm: "milestone",
        objectiveEnabled: true,
        keyResultEnabled: true,
        workingDays: [1, 2, 3, 4, 5],
        workingStartMinute: 540,
        workingEndMinute: 1020,
      },
    } as ReturnType<typeof useWorkspaceSettings>);
    queryClient.setQueryData(teamKeys.lists("first"), [
      { id: "sales", storyTerm: "deal" },
      { id: "engineering", storyTerm: null },
    ]);
  });

  it("uses the route's team work name while preserving other workspace terms", () => {
    const { result } = renderHook(() => useTerminology(), { wrapper });
    expect(
      result.current.getTermDisplay("storyTerm", {
        variant: "plural",
        capitalize: true,
      }),
    ).toBe("Deals");
    expect(result.current.getTermDisplay("objectiveTerm")).toBe("project");
    expect(
      result.current.getTermDisplay("sprintTerm", { variant: "plural" }),
    ).toBe("cycles");
  });

  it("uses an explicit story team and supports global workspace fallback", () => {
    const { result, rerender } = renderHook(
      ({ teamId }: { teamId: string | null }) => useTerminology(teamId),
      { initialProps: { teamId: "engineering" as string | null }, wrapper },
    );
    expect(result.current.getTermDisplay("storyTerm")).toBe("task");
    rerender({ teamId: "sales" });
    expect(result.current.getTermDisplay("storyTerm")).toBe("deal");
    rerender({ teamId: null });
    expect(result.current.getTermDisplay("storyTerm")).toBe("task");
    expect(queryClient.getQueryCache().getAll()).toHaveLength(1);
  });

  it("retains workspace defaults when a team is absent or unavailable", () => {
    jest.mocked(useParams).mockReturnValue({});
    const { result } = renderHook(() => useTerminology(), { wrapper });
    expect(
      result.current.getTermDisplay("storyTerm", { variant: "plural" }),
    ).toBe("tasks");
  });

  it("updates a team override from the canonical cache without a second query owner", () => {
    const { result } = renderHook(() => useTerminology(), { wrapper });
    expect(result.current.getTermDisplay("storyTerm")).toBe("deal");
    act(() => {
      queryClient.setQueryData(teamKeys.lists("first"), [
        { id: "sales", storyTerm: "issue" },
      ]);
    });
    expect(result.current.getTermDisplay("storyTerm")).toBe("issue");
    expect(queryClient.getQueryCache().getAll()).toHaveLength(1);
  });
});
