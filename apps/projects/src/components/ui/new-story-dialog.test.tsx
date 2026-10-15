/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useImperativeHandle, useState } from "react";
import type { ReactNode } from "react";
import { CreationPropertySlotsProvider } from "@/shared/story/creation-property-slots";
import type { CreationPropertiesProps } from "@/shared/story/creation-property-slots";
import { NewStoryDialog } from "./new-story-dialog";

const mockSetIsOpen = jest.fn();
const mockSaveDraft = jest.fn().mockResolvedValue(undefined);
const mockDraft = {
  supported: false,
  ready: true,
  recovered: null as { version: number; form: Record<string, unknown> } | null,
  status: "",
  error: null,
  save: mockSaveDraft,
  discard: jest.fn().mockResolvedValue(undefined),
  retry: jest.fn(),
  acceptRecovered: () => {
    mockDraft.recovered = null;
  },
};
jest.mock("@/shell/desktop/use-desktop-draft", () => ({
  useDesktopDraft: () => mockDraft,
}));

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
    customFields,
    draftNotice,
  }: {
    customFields: ReactNode;
    draftNotice: ReactNode;
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
      {draftNotice}
      {customFields}
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
    mockSaveDraft.mockClear();
    mockDraft.supported = false;
    mockDraft.recovered = null;
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

const DraftProperties = ({
  controllerRef,
  onValuesChange,
  disabled,
}: CreationPropertiesProps) => {
  const [values, setValues] = useState<Record<string, string | null>>({});
  const update = (next: Record<string, string | null>) => {
    setValues(next);
    onValuesChange?.(next);
  };
  useImperativeHandle(controllerRef, () => ({
    isPending: false,
    isError: false,
    prepareValues: () => [],
    reset: () => {
      update({});
    },
    setValues: update,
  }));
  return (
    <>
      <button
        disabled={disabled}
        onClick={() => {
          update({
            amount: "9007199254740993.29",
            note: "  unfinished text  ",
            cleared: null,
          });
        }}
        type="button"
      >
        Edit custom fields
      </button>
      <output data-testid="custom-values">{JSON.stringify(values)}</output>
    </>
  );
};
const draftSlots = { Properties: DraftProperties, TemplatePicker: () => null };
it("durably captures and restores custom fields even before a task title is entered", async () => {
  mockSaveDraft.mockClear();
  mockDraft.supported = true;
  mockDraft.recovered = null;
  const content = (
    <CreationPropertySlotsProvider slots={draftSlots}>
      <NewStoryDialog isOpen setIsOpen={mockSetIsOpen} />
    </CreationPropertySlotsProvider>
  );
  const first = render(content);
  fireEvent.click(screen.getByRole("button", { name: "Edit custom fields" }));
  await waitFor(() => {
    expect(mockSaveDraft).toHaveBeenLastCalledWith(
      expect.objectContaining({
        form: expect.objectContaining({
          title: "",
          customFieldValues: [
            { fieldId: "amount", value: "9007199254740993.29" },
            { fieldId: "note", value: "  unfinished text  " },
            { fieldId: "cleared", value: null },
          ],
        }),
      }),
    );
  });
  const saved = mockSaveDraft.mock.calls.at(-1)[0];
  first.unmount();
  mockDraft.recovered = saved;
  render(content);
  expect(
    screen.getByRole("button", { name: "Edit custom fields" }),
  ).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Restore draft" }));
  expect(JSON.parse(screen.getByTestId("custom-values").textContent)).toEqual({
    amount: "9007199254740993.29",
    note: "  unfinished text  ",
    cleared: null,
  });
  expect(
    screen.getByRole("button", { name: "Edit custom fields" }),
  ).toBeEnabled();
});
