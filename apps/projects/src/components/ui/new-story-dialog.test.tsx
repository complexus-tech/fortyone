/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { NewStoryDialog } from "./new-story-dialog";

const mockSetIsOpen = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: jest.fn() }),
}));

jest.mock("@tanstack/react-query", () => ({
  useQueryClient: () => ({ invalidateQueries: jest.fn() }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "current-user" } } }),
}));

jest.mock("@/hooks", () => ({
  useFeatures: () => ({ objectiveEnabled: false }),
  useLocalStorage: <T,>(_key: string, initialValue: T) => [
    initialValue,
    jest.fn(),
  ],
  useSprintsEnabled: () => false,
  useTerminology: () => ({
    getTermDisplay: (term: string, options?: { capitalize?: boolean }) =>
      options?.capitalize ? term[0].toUpperCase() + term.slice(1) : term,
  }),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => path,
    workspaceSlug: "workspace-1",
  }),
}));

jest.mock("@/hooks/debounce", () => ({
  useDebouncedCallback: (callback: (value: string) => void) => ({
    callback,
    cancel: jest.fn(),
  }),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({
    data: [
      { id: "status-1", isDefault: true, teamId: "team-1" },
      { id: "status-2", teamId: "team-1" },
    ],
  }),
}));

jest.mock("@/lib/hooks/labels", () => ({
  useLabels: () => ({ data: [] }),
}));

jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => ({ data: null, isLoading: false }),
  useMembers: () => ({ data: [] }),
}));

jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({ data: { autoAssignSelf: true } }),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: () => ({
    getLimit: () => 100,
    hasFeature: () => false,
    tier: "free",
  }),
}));

jest.mock("@/lib/hooks/figma", () => ({
  useLinkFigmaStory: () => ({ mutateAsync: jest.fn() }),
}));

jest.mock("@/modules/teams/hooks/teams", () => ({
  useJoinedTeams: () => ({
    data: [
      {
        code: "ENG",
        color: "#000000",
        id: "team-1",
        name: "Engineering",
      },
    ],
  }),
}));

jest.mock("@/modules/teams/hooks/use-team-settings", () => ({
  useTeamSettings: () => ({
    data: { estimationSettings: { scheme: "fibonacci" } },
  }),
}));

jest.mock("@/modules/objectives/hooks/use-objectives", () => ({
  useTeamObjectives: () => ({ data: [] }),
}));

jest.mock("@/modules/objectives/hooks", () => ({
  useKeyResults: () => ({ data: [] }),
}));

jest.mock("@/modules/sprints/hooks/team-sprints", () => ({
  useTeamSprints: () => ({ data: [] }),
}));

jest.mock("@/modules/stories/hooks/total-stories", () => ({
  useTotalStories: () => ({ data: 0 }),
}));

jest.mock("@/modules/story/hooks/create-mutation", () => ({
  useCreateStoryMutation: () => ({ mutateAsync: jest.fn() }),
}));

jest.mock("@/modules/story/hooks/use-story-description-media", () => ({
  useStoryDescriptionMedia: () => ({
    cancelStagedUploads: jest.fn(),
    finalizeStagedMedia: jest.fn(),
    handleMediaFiles: jest.fn(),
    inputRef: { current: null },
    openMediaPicker: jest.fn(),
    resetForNextStory: jest.fn(),
  }),
}));

jest.mock("@/modules/search/hooks/use-similar-stories", () => ({
  useSimilarStories: () => ({ data: [] }),
}));

jest.mock("./use-new-story-dialog-editors", () => ({
  useNewStoryDialogEditors: () => ({
    descriptionEditor: null,
    titleEditor: null,
  }),
}));

jest.mock("./use-new-story-dialog-creation", () => ({
  useNewStoryDialogCreation: () => ({
    handleCreateStory: jest.fn(),
    isCreating: false,
  }),
}));

jest.mock("./use-new-story-dialog-lifecycle", () => ({
  useMayaAutoScheduling: jest.fn(),
  useNewStoryDialogLifecycle: jest.fn(),
}));

jest.mock("./new-story-dialog-limit-guard", () => ({
  NewStoryDialogLimitGuard: ({ children }: { children: ReactNode }) => (
    <>{children}</>
  ),
}));

jest.mock("./new-story-dialog-content", () => ({
  NewStoryDialogContent: ({
    onOpenChange,
    storyForm,
  }: {
    onOpenChange: (open: boolean) => void;
    storyForm: unknown;
  }) => (
    <>
      <button
        onClick={() => {
          onOpenChange(false);
        }}
        type="button"
      >
        Close dialog
      </button>
      <output data-testid="story-form">{JSON.stringify(storyForm)}</output>
    </>
  ),
}));

const readStoryForm = () =>
  JSON.parse(screen.getByTestId("story-form").textContent) as {
    assigneeId?: string | null;
    priority?: string;
    statusId?: string;
  };

describe("NewStoryDialog", () => {
  beforeEach(() => {
    mockSetIsOpen.mockReset();
  });

  it("reinitializes the mounted draft when the caller changes creation context", async () => {
    const { rerender } = render(
      <NewStoryDialog
        isOpen
        priority="Low"
        setIsOpen={mockSetIsOpen}
        statusId="status-1"
      />,
    );

    expect(readStoryForm()).toMatchObject({
      assigneeId: "current-user",
      priority: "Low",
      statusId: "status-1",
    });

    rerender(
      <NewStoryDialog
        assigneeId="assignee-2"
        isOpen
        priority="High"
        setIsOpen={mockSetIsOpen}
        statusId="status-2"
      />,
    );

    await waitFor(() => {
      expect(readStoryForm()).toMatchObject({
        assigneeId: "assignee-2",
        priority: "High",
        statusId: "status-2",
      });
    });
  });

  it("propagates a dialog close through the caller-owned open state", () => {
    render(<NewStoryDialog isOpen setIsOpen={mockSetIsOpen} />);

    fireEvent.click(screen.getByRole("button", { name: "Close dialog" }));

    expect(mockSetIsOpen).toHaveBeenCalledWith(false);
  });
});
