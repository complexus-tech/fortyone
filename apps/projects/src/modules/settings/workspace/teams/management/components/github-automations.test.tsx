import { fireEvent, render, screen } from "@testing-library/react";
import type { GitHubTeamSettings } from "@/modules/settings/workspace/integrations/github/types";
import { GitHubAutomations } from "./github-automations";

const mockMutate = jest.fn();
const mockSettingsQuery: {
  data: GitHubTeamSettings | undefined;
  isPending: boolean;
  isError: boolean;
} = { data: undefined, isPending: false, isError: false };
const mockStatusesQuery: {
  data: { id: string; name: string }[] | undefined;
  isPending: boolean;
  isError: boolean;
} = { data: undefined, isPending: false, isError: false };
const mockMutation = { mutate: mockMutate, isPending: false };

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team-id" }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "task" }),
}));
jest.mock("@/lib/hooks/github/use-team-settings", () => ({
  useGitHubTeamSettings: () => mockSettingsQuery,
}));
jest.mock("@/lib/hooks/github/use-update-team-settings", () => ({
  useUpdateGitHubTeamSettings: () => mockMutation,
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useTeamStatuses: () => mockStatusesQuery,
}));
jest.mock("@/components/ui/story-status-icon", () => ({
  StoryStatusIcon: () => <span aria-hidden />,
}));

const scrollIntoViewDescriptor = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
);

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: jest.fn(),
  });
});

afterAll(() => {
  if (scrollIntoViewDescriptor) {
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollIntoViewDescriptor,
    );
  } else {
    Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
  }
});

beforeEach(() => {
  mockMutate.mockClear();
  mockMutation.isPending = false;
  Object.assign(mockSettingsQuery, {
    data: {
      teamId: "team-id",
      rules: [
        {
          id: "open-rule",
          eventKey: "pr_open",
          targetStatusId: "review",
          baseBranchPattern: "main",
          isActive: true,
        },
        {
          id: "merge-rule",
          eventKey: "pr_merge",
          targetStatusId: "done",
          baseBranchPattern: "release/*",
          isActive: false,
        },
      ],
    },
    isPending: false,
    isError: false,
  });
  Object.assign(mockStatusesQuery, {
    data: [
      { id: "review", name: "In review" },
      { id: "done", name: "Done" },
    ],
    isPending: false,
    isError: false,
  });
});

it.each(["settings", "statuses"] as const)(
  "prevents edits while %s are loading and keeps trigger text visible",
  (queryName) => {
    const query =
      queryName === "settings" ? mockSettingsQuery : mockStatusesQuery;
    query.data = undefined;
    query.isPending = true;
    render(<GitHubAutomations />);

    for (const trigger of screen.getAllByRole("combobox")) {
      expect(trigger).toBeDisabled();
      expect(trigger).toHaveTextContent("Loading settings...");
    }
    expect(mockMutate).not.toHaveBeenCalled();
  },
);

it("shows an error and prevents edits if statuses fail to load", () => {
  mockStatusesQuery.data = undefined;
  mockStatusesQuery.isError = true;
  render(<GitHubAutomations />);

  expect(screen.getByRole("alert")).toHaveTextContent("Could not load");
  for (const trigger of screen.getAllByRole("combobox")) {
    expect(trigger).toBeDisabled();
    expect(trigger).toHaveTextContent("Settings unavailable");
  }
});

it("labels a saved status that no longer exists instead of rendering a blank value", () => {
  mockStatusesQuery.data = [];
  render(<GitHubAutomations />);
  expect(
    screen.getByRole("combobox", { name: "On PR open, move to..." }),
  ).toHaveTextContent("Unavailable status");
});

it("prevents another update while saving and retains the selected label", () => {
  mockMutation.isPending = true;
  render(<GitHubAutomations />);
  expect(
    screen.getByRole("combobox", { name: "On PR open, move to..." }),
  ).toHaveTextContent("In review");
  for (const trigger of screen.getAllByRole("combobox")) {
    expect(trigger).toBeDisabled();
  }
});

it("clears only the chosen target and preserves other saved rule settings", () => {
  render(<GitHubAutomations />);
  fireEvent.keyDown(
    screen.getByRole("combobox", { name: "On PR open, move to..." }),
    { key: "ArrowDown" },
  );
  fireEvent.click(screen.getByRole("option", { name: "No action" }));

  expect(mockMutate).toHaveBeenCalledTimes(1);
  expect(mockMutate).toHaveBeenCalledWith({
    rules: expect.arrayContaining([
      {
        eventKey: "pr_open",
        targetStatusId: null,
        baseBranchPattern: "main",
        isActive: true,
      },
      {
        eventKey: "pr_merge",
        targetStatusId: "done",
        baseBranchPattern: "release/*",
        isActive: false,
      },
    ]),
  });
});
