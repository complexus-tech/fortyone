/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { PropsWithChildren } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import {
  useFeatures,
  useMediaQuery,
  useSprintsEnabled,
  useTerminology,
  useUserRole,
} from "@/hooks";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { useMayaAssignee, useMembers } from "@/lib/hooks/members";
import { useStatuses } from "@/lib/hooks/statuses";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useKeyResults } from "@/modules/objectives/hooks";
import { useObjective } from "@/modules/objectives/hooks/use-objective";
import { useSprint } from "@/modules/sprints/hooks/sprint-details";
import { useStoryById } from "@/modules/story/hooks/story";
import { useUpdateStoryMutation } from "../hooks/update-mutation";
import { Options } from "./options";

jest.mock("ui", () => ({
  Badge: function MockBadge({ children }: PropsWithChildren) {
    return <span>{children}</span>;
  },
  Box: function MockBox({
    children,
    className,
  }: PropsWithChildren<{ className?: string }>) {
    return <div className={className}>{children}</div>;
  },
  Container: function MockContainer({
    children,
    className,
  }: PropsWithChildren<{ className?: string }>) {
    return <div className={className}>{children}</div>;
  },
  Divider: function MockDivider({ className }: { className?: string }) {
    return <hr className={className} />;
  },
  Text: function MockText({ children }: PropsWithChildren) {
    return <span>{children}</span>;
  },
}));

jest.mock("@/components/ui/confirm-dialog", () => ({
  ConfirmDialog: function MockConfirmDialog({
    cancelText,
    confirmText,
    isOpen,
    onCancel,
    onConfirm,
    title,
  }: {
    cancelText: string;
    confirmText: string;
    isOpen: boolean;
    onCancel: () => void;
    onConfirm: () => void;
    title: string;
  }) {
    if (!isOpen) return null;

    return (
      <div aria-label={title} role="dialog">
        <button onClick={onCancel} type="button">
          {cancelText}
        </button>
        <button onClick={onConfirm} type="button">
          {confirmText}
        </button>
      </div>
    );
  },
}));

jest.mock("@/hooks", () => ({
  useFeatures: jest.fn(),
  useMediaQuery: jest.fn(),
  useSprintsEnabled: jest.fn(),
  useTerminology: jest.fn(),
  useUserRole: jest.fn(),
}));

jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: jest.fn(),
}));

jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: jest.fn(),
  useMembers: jest.fn(),
}));

jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: jest.fn(),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  useSubscriptionFeatures: jest.fn(),
}));

jest.mock("@/modules/objectives/hooks", () => ({
  useKeyResults: jest.fn(),
}));

jest.mock("@/modules/objectives/hooks/use-objective", () => ({
  useObjective: jest.fn(),
}));

jest.mock("@/modules/sprints/hooks/sprint-details", () => ({
  useSprint: jest.fn(),
}));

jest.mock("@/modules/story/hooks/story", () => ({
  useStoryById: jest.fn(),
}));

jest.mock("../hooks/update-mutation", () => ({
  useUpdateStoryMutation: jest.fn(),
}));

jest.mock("./add-links", () => ({
  AddLinks: function MockAddLinks() {
    return null;
  },
}));

jest.mock("./options-header", () => ({
  OptionsHeader: function MockOptionsHeader() {
    return null;
  },
}));

jest.mock("./story-options/core-options", () => ({
  CoreOptions: function MockCoreOptions({
    disabled,
    onUpdate,
  }: {
    disabled: boolean;
    onUpdate: (payload: { statusId: string }) => void;
  }) {
    return (
      <button
        disabled={disabled}
        onClick={() => {
          onUpdate({ statusId: "status-completed" });
        }}
        type="button"
      >
        Complete story
      </button>
    );
  },
}));

jest.mock("./story-options/date-options", () => ({
  DateOptions: function MockDateOptions() {
    return null;
  },
}));

jest.mock("./story-options/labels-option", () => ({
  LabelsOption: function MockLabelsOption() {
    return null;
  },
}));

jest.mock("./story-options/planning-options", () => ({
  PlanningOptions: function MockPlanningOptions() {
    return null;
  },
}));

const mockedUseFeatures = jest.mocked(useFeatures);
const mockedUseMediaQuery = jest.mocked(useMediaQuery);
const mockedUseSprintsEnabled = jest.mocked(useSprintsEnabled);
const mockedUseTerminology = jest.mocked(useTerminology);
const mockedUseUserRole = jest.mocked(useUserRole);
const mockedUseIsAdminOrOwner = jest.mocked(useIsAdminOrOwner);
const mockedUseMayaAssignee = jest.mocked(useMayaAssignee);
const mockedUseMembers = jest.mocked(useMembers);
const mockedUseStatuses = jest.mocked(useStatuses);
const mockedUseSubscriptionFeatures = jest.mocked(useSubscriptionFeatures);
const mockedUseKeyResults = jest.mocked(useKeyResults);
const mockedUseObjective = jest.mocked(useObjective);
const mockedUseSprint = jest.mocked(useSprint);
const mockedUseStoryById = jest.mocked(useStoryById);
const mockedUseUpdateStoryMutation = jest.mocked(useUpdateStoryMutation);

const statuses = [
  { category: "started", id: "status-started", name: "In progress" },
  { category: "backlog", id: "status-backlog", name: "Backlog" },
  { category: "completed", id: "status-completed", name: "Done" },
];

const story = {
  assigneeId: null,
  deletedAt: null,
  endDate: null,
  id: "story-parent",
  keyResultId: null,
  labels: [],
  objectiveId: null,
  reporterId: "reporter-1",
  sprintId: null,
  startDate: null,
  statusId: "status-started",
  subStories: [
    { id: "child-started", statusId: "status-started" },
    { id: "child-backlog", statusId: "status-backlog" },
    { id: "child-completed", statusId: "status-completed" },
  ],
  teamId: "team-1",
};

describe("Options", () => {
  const mutate = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    mockedUseFeatures.mockReturnValue({
      objectiveEnabled: true,
    } as ReturnType<typeof useFeatures>);
    mockedUseMediaQuery.mockReturnValue(false);
    mockedUseSprintsEnabled.mockReturnValue(true);
    mockedUseTerminology.mockReturnValue({
      getTermDisplay: () => "story",
    } as ReturnType<typeof useTerminology>);
    mockedUseUserRole.mockReturnValue({
      userRole: "member",
    } as ReturnType<typeof useUserRole>);
    mockedUseIsAdminOrOwner.mockReturnValue({
      isAdminOrOwner: false,
    } as ReturnType<typeof useIsAdminOrOwner>);
    mockedUseMayaAssignee.mockReturnValue({
      data: undefined,
    } as ReturnType<typeof useMayaAssignee>);
    mockedUseMembers.mockReturnValue({
      data: [],
    } as unknown as ReturnType<typeof useMembers>);
    mockedUseStatuses.mockReturnValue({
      data: statuses,
    } as unknown as ReturnType<typeof useStatuses>);
    mockedUseSubscriptionFeatures.mockReturnValue({
      hasFeature: () => false,
    } as unknown as ReturnType<typeof useSubscriptionFeatures>);
    mockedUseKeyResults.mockReturnValue({
      data: [],
    } as unknown as ReturnType<typeof useKeyResults>);
    mockedUseObjective.mockReturnValue({
      data: undefined,
    } as ReturnType<typeof useObjective>);
    mockedUseSprint.mockReturnValue({
      data: undefined,
    } as ReturnType<typeof useSprint>);
    mockedUseStoryById.mockReturnValue({
      data: story,
    } as unknown as ReturnType<typeof useStoryById>);
    mockedUseUpdateStoryMutation.mockReturnValue({
      mutate,
    } as unknown as ReturnType<typeof useUpdateStoryMutation>);
  });

  it("defers completion and updates only unfinished children after confirmation", () => {
    render(
      <Options
        isNotifications={false}
        storyId="story-parent"
        variant="inline"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Complete story" }));

    expect(mutate).not.toHaveBeenCalled();
    expect(screen.getByRole("dialog")).toBeVisible();

    fireEvent.click(screen.getByRole("button", { name: "Yes, mark as done" }));

    expect(mutate.mock.calls).toEqual([
      [
        {
          payload: { statusId: "status-completed" },
          storyId: "story-parent",
        },
      ],
      [
        {
          payload: { statusId: "status-completed" },
          storyId: "child-started",
        },
      ],
      [
        {
          payload: { statusId: "status-completed" },
          storyId: "child-backlog",
        },
      ],
    ]);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("updates only the parent when child completion is declined", () => {
    render(
      <Options
        isNotifications={false}
        storyId="story-parent"
        variant="inline"
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Complete story" }));
    fireEvent.click(screen.getByRole("button", { name: "No, leave as is" }));

    expect(mutate).toHaveBeenCalledTimes(1);
    expect(mutate).toHaveBeenCalledWith({
      payload: { statusId: "status-completed" },
      storyId: "story-parent",
    });
  });

  it("disables property editing for guests", () => {
    mockedUseUserRole.mockReturnValue({
      userRole: "guest",
    } as ReturnType<typeof useUserRole>);

    render(
      <Options
        isNotifications={false}
        storyId="story-parent"
        variant="inline"
      />,
    );

    expect(
      screen.getByRole("button", { name: "Complete story" }),
    ).toBeDisabled();
  });
});
