import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useTerminology } from "@/hooks";
import { OptionsHeader } from "./options-header";

jest.mock("@/hooks", () => ({
  useCopyToClipboard: () => [null, jest.fn()],
  useTerminology: jest.fn(() => ({ getTermDisplay: () => "story" })),
  useUserRole: () => ({ userRole: "member" }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => path,
    workspaceSlug: "acme",
  }),
}));
jest.mock("@/modules/story/hooks/story", () => ({
  useStoryById: () => ({
    data: { id: "story", teamId: "team", title: "Launch", sequenceId: 276 },
  }),
}));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [{ id: "team", code: "TM" }] }),
}));
jest.mock("@/lib/hooks/profile", () => ({ useProfile: () => ({ data: {} }) }));
jest.mock("@/lib/hooks/users/preferences", () => ({
  useAutomationPreferences: () => ({ data: {} }),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/github", () => ({
  useGitHubIntegration: () => ({ data: null }),
}));
jest.mock("@/components/shared", () => ({ MobileMenuButton: () => null }));
jest.mock("@/shared/favorites", () => ({ FavoriteButton: () => null }));
jest.mock("../hooks/update-mutation", () => ({
  useUpdateStoryMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("./story-actions-menu", () => ({
  StoryActionsMenu: () => <button type="button">Story actions</button>,
}));

afterEach(cleanup);
beforeEach(() => {
  jest
    .mocked(useTerminology)
    .mockReturnValue({ getTermDisplay: () => "story" });
});

it("places the side-panel close control next to the existing story actions", () => {
  const onClose = jest.fn();
  render(<OptionsHeader isAdminOrOwner onClose={onClose} storyId="story" />);
  const buttons = screen.getAllByRole("button");
  const close = screen.getByRole("button", { name: "Close story" });
  expect(buttons.at(-1)).toBe(close);
  expect(buttons.at(-2)).toHaveAccessibleName("Story actions");
  expect(
    screen.getByRole("button", { name: "Copy git branch name" }),
  ).toBeVisible();
  expect(screen.getByText("TM-276")).toBeVisible();
  expect(close).toHaveFocus();
  fireEvent.click(close);
  expect(onClose).toHaveBeenCalledTimes(1);
});

it("leaves standalone story headers without a dismissal control", () => {
  render(<OptionsHeader isAdminOrOwner storyId="story" />);
  expect(
    screen.queryByRole("button", { name: "Close story" }),
  ).not.toBeInTheDocument();
});

it("uses the story team's terminology for its close control", () => {
  jest.mocked(useTerminology).mockReturnValue({ getTermDisplay: () => "task" });
  render(<OptionsHeader isAdminOrOwner onClose={jest.fn()} storyId="story" />);
  expect(useTerminology).toHaveBeenCalledWith("team");
  expect(screen.getByRole("button", { name: "Close task" })).toBeVisible();
  expect(
    screen.queryByRole("button", { name: "Close story" }),
  ).not.toBeInTheDocument();
});
