import type { ComponentProps } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { NewStoryDialogContent } from "./new-story-dialog-content";

jest.mock("./new-story-dialog-header", () => ({
  NewStoryDialogHeader: () => <h2>New task</h2>,
}));
jest.mock("./new-story-dialog-fields", () => ({
  NewStoryDialogFields: () => null,
}));
jest.mock("./new-story-figma-source", () => ({
  NewStoryFigmaSource: () => null,
}));
jest.mock("./new-story-dialog-similar-stories", () => ({
  NewStoryDialogSimilarStories: () => null,
}));
jest.mock("@/lib/tiptap/rich-text-table-menu", () => ({
  RichTextTableMenu: () => null,
}));
const props: ComponentProps<typeof NewStoryDialogContent> = {
  canUseBackgroundMaya: true,
  createMore: false,
  currentTeam: null,
  currentTeamId: "team",
  deadlineSourceRef: { current: "unset" },
  descriptionEditor: null,
  dispatch: jest.fn(),
  estimateScheme: "points",
  figmaArtifacts: [],
  isCreating: false,
  isExpanded: false,
  isMayaAssigned: false,
  isMayaAssigneeLoading: true,
  isOpen: true,
  mediaInputRef: { current: null },
  member: undefined,
  mayaAssigneeId: undefined,
  sprintName: undefined,
  members: [],
  objectiveTerm: "Objective",
  onActiveTeamChange: jest.fn(),
  onCreate: jest.fn(),
  onCreateMoreChange: jest.fn(),
  onFigmaArtifactsChange: jest.fn(),
  onMediaFiles: jest.fn(),
  onOpenChange: jest.fn(),
  onSimilarStorySelect: jest.fn(),
  onToggleExpanded: jest.fn(),
  selectedLabels: [],
  showObjectives: false,
  showSprints: false,
  similarStories: [],
  sprintTerm: "Sprint",
  statuses: [],
  storyForm: { title: "", priority: "No Priority" },
  storyTerm: "task",
  storyTermPlural: "tasks",
  strategyLinkLabel: undefined,
  teamStatuses: [],
  teams: [],
  titleEditor: null,
};
it("keeps ordinary task submission available while optional Maya metadata loads", () => {
  render(<NewStoryDialogContent {...props} />);
  const create = screen.getByRole("button", { name: "Create task" });
  expect(create).toBeEnabled();
  fireEvent.click(create);
  expect(props.onCreate).toHaveBeenCalledTimes(1);
  expect(screen.queryByText("Preparing Maya...")).not.toBeInTheDocument();
});
it("shows preparation only for work assigned to Maya", () => {
  render(<NewStoryDialogContent {...props} isMayaAssigned />);
  expect(
    screen.getByRole("button", { name: "Preparing Maya..." }),
  ).toBeDisabled();
});
