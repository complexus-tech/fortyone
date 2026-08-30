/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { useSession } from "@/lib/auth/client";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { storyKeys } from "@/modules/stories/constants";
import type { Sprint } from "@/modules/sprints/types";
import { SprintRow } from "./row";

jest.mock("@/lib/auth/client", () => ({
  useSession: jest.fn(),
}));

jest.mock("@/hooks", () => ({
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("@/components/ui", () => ({
  StoryStatusIcon: () => <span />,
}));

jest.mock("@/components/ui/row-wrapper", () => ({
  RowWrapper: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

jest.mock("@/modules/stories/queries/get-stories", () => ({
  getStories: jest.fn(),
}));

jest.mock("icons", () => ({
  ArrowRightIcon: () => <span />,
  CalendarIcon: () => <span />,
  SprintsIcon: () => <span />,
}));

jest.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    href,
    onMouseEnter,
  }: {
    children: ReactNode;
    href: string;
    onMouseEnter?: () => void;
  }) => (
    <a href={href} onMouseEnter={onMouseEnter}>
      {children}
    </a>
  ),
}));

jest.mock("ui", () => ({
  Badge: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  ProgressBar: () => <span />,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  Tooltip: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

const sprint: Sprint = {
  createdAt: "2026-08-01T00:00:00.000Z",
  endDate: "2026-08-31T00:00:00.000Z",
  goal: "Ship the sprint cache fix",
  id: "sprint-1",
  name: "Sprint 1",
  objectiveId: "objective-1",
  scheduleManagedByAutomation: false,
  startDate: "2026-08-01T00:00:00.000Z",
  stats: {
    backlog: 1,
    cancelled: 0,
    completed: 2,
    started: 1,
    total: 4,
    unstarted: 0,
  },
  teamId: "team-1",
  updatedAt: "2026-08-01T00:00:00.000Z",
  workspaceId: "workspace-1",
};

const mockUseSession = jest.mocked(useSession);
const mockUseTerminology = jest.mocked(useTerminology);
const mockUseWorkspacePath = jest.mocked(useWorkspacePath);

describe("SprintRow", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseSession.mockReturnValue({
      data: { token: "session-token" },
    } as unknown as ReturnType<typeof useSession>);
    mockUseTerminology.mockReturnValue({
      getTermDisplay: () => "Sprint",
    } as ReturnType<typeof useTerminology>);
    mockUseWorkspacePath.mockReturnValue({
      withWorkspace: (path: string) => `/acme${path}`,
      workspaceSlug: "acme",
    } as ReturnType<typeof useWorkspacePath>);
  });

  it("prefetches sprint stories through the surrounding query client", () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });
    const prefetchQuery = jest
      .spyOn(queryClient, "prefetchQuery")
      .mockResolvedValue(undefined);

    render(
      <QueryClientProvider client={queryClient}>
        <SprintRow {...sprint} />
      </QueryClientProvider>,
    );

    fireEvent.mouseEnter(screen.getByRole("link", { name: /Sprint 1/ }));

    expect(prefetchQuery).toHaveBeenCalledWith({
      queryFn: expect.any(Function),
      queryKey: storyKeys.sprint("acme", "sprint-1"),
      staleTime: 180_000,
    });
  });
});
