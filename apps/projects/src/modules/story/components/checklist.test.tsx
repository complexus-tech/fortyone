/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import type { DetailedStory } from "../types";
import type { StoryChecklistItem } from "../types/checklist";
import { useChecklistMutation, useStoryChecklist } from "../hooks/checklist";
import { StoryChecklist } from "./checklist";

let mockUserRole = "member";
const mockMutate = jest.fn();
const mockRefetch = jest.fn();

jest.mock("@/hooks", () => ({
  useUserRole: () => ({ userRole: mockUserRole }),
  useTerminology: () => ({ getTermDisplay: () => "story" }),
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("../hooks/checklist", () => ({
  useStoryChecklist: jest.fn(),
  useChecklistMutation: jest.fn(),
}));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({
    data: [{ id: "team-1", code: "ENG", name: "Engineering" }],
  }),
}));
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [{ id: "todo", name: "To do", color: "#fff" }] }),
}));

const story = {
  id: "story-1",
  teamId: "team-1",
  deletedAt: null,
} as DetailedStory;
const makeItem = (
  overrides: Partial<StoryChecklistItem> = {},
): StoryChecklistItem => ({
  id: "item-1",
  storyId: story.id,
  title: "Review design",
  completed: false,
  position: 0,
  convertedAt: null,
  convertedStory: null,
  createdAt: "2026-10-13T00:00:00Z",
  updatedAt: "2026-10-13T00:00:00Z",
  ...overrides,
});

const setQuery = (items: StoryChecklistItem[], overrides = {}) => {
  jest.mocked(useStoryChecklist).mockReturnValue({
    data: items,
    isLoading: false,
    isError: false,
    isFetching: false,
    refetch: mockRefetch,
    ...overrides,
  } as unknown as ReturnType<typeof useStoryChecklist>);
};

const openMenu = (name: string) => {
  fireEvent.pointerDown(screen.getByRole("button", { name }), {
    button: 0,
    ctrlKey: false,
  });
};

beforeEach(() => {
  jest.clearAllMocks();
  mockUserRole = "member";
  setQuery([]);
  jest.mocked(useChecklistMutation).mockReturnValue({
    mutate: mockMutate,
    isPending: false,
  } as unknown as ReturnType<typeof useChecklistMutation>);
});

describe("StoryChecklist", () => {
  it("creates trimmed items with an explicit form and retains the draft until success", () => {
    render(<StoryChecklist story={story} />);
    fireEvent.click(screen.getByRole("button", { name: "Add checklist" }));
    const input = screen.getByRole("textbox", { name: "New checklist item" });
    expect(input).toHaveFocus();
    expect(input).toHaveAttribute("maxlength", "255");
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
    fireEvent.change(input, { target: { value: "  Review design  " } });
    fireEvent.click(screen.getByRole("button", { name: "Create" }));
    expect(mockMutate).toHaveBeenCalledWith(
      { type: "create", title: "Review design" },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    expect(input).toHaveValue("  Review design  ");
  });

  it("cancels a draft with Escape without changing the story", () => {
    render(<StoryChecklist story={story} />);
    fireEvent.click(screen.getByRole("button", { name: "Add checklist" }));
    fireEvent.keyDown(screen.getByRole("textbox"), { key: "Escape" });
    expect(screen.getByRole("button", { name: "Add checklist" })).toBeVisible();
    expect(mockMutate).not.toHaveBeenCalled();
  });

  it("tracks checklist completion independently and exposes keyboard sorting handles", () => {
    setQuery([
      makeItem(),
      makeItem({
        id: "item-2",
        title: "Test checkout",
        completed: true,
        position: 1,
      }),
    ]);
    render(<StoryChecklist story={story} />);
    expect(
      screen.getByLabelText("1 of 2 checklist items complete"),
    ).toHaveTextContent("1/2");
    fireEvent.click(
      screen.getByRole("checkbox", { name: "Complete Review design" }),
    );
    expect(mockMutate).toHaveBeenCalledWith(
      { type: "update", itemId: "item-1", updates: { completed: true } },
      { onSuccess: undefined },
    );
    expect(
      screen.getByRole("button", { name: "Reorder Review design" }),
    ).toHaveAttribute("tabindex", "0");
    expect(
      screen.getByRole("button", { name: "Reorder Review design" }),
    ).toHaveAttribute("aria-describedby");
  });

  it("renames an item inline and preserves its draft while the request is unresolved", () => {
    setQuery([makeItem()]);
    render(<StoryChecklist story={story} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit Review design" }));
    fireEvent.change(
      screen.getByRole("textbox", { name: "Edit checklist item" }),
      { target: { value: "Review final design" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(mockMutate).toHaveBeenCalledWith(
      {
        type: "update",
        itemId: "item-1",
        updates: { title: "Review final design" },
      },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    expect(screen.getByRole("textbox")).toHaveValue("Review final design");
  });

  it("converts an item through its menu without creating a sub-story", () => {
    setQuery([makeItem()]);
    render(<StoryChecklist story={story} />);
    openMenu("Actions for Review design");
    fireEvent.click(screen.getByRole("menuitem", { name: "Convert to story" }));
    expect(mockMutate).toHaveBeenCalledWith({
      type: "convert",
      itemId: "item-1",
    });
  });

  it("retains converted items as completed links and prevents repeat conversion", () => {
    setQuery([
      makeItem({
        completed: true,
        convertedAt: "2026-10-13T01:00:00Z",
        convertedStory: {
          id: "converted-story",
          sequenceId: 231,
          teamId: "team-1",
          title: "Review design",
          statusId: "todo",
        },
      }),
    ]);
    render(<StoryChecklist story={story} />);
    expect(screen.getByRole("checkbox")).toBeChecked();
    expect(screen.getByRole("checkbox")).toBeDisabled();
    expect(
      screen.getByRole("link", { name: "ENG-231: Review design" }),
    ).toHaveAttribute("href", "/acme/work/ENG-231");
    openMenu("Actions for Review design");
    expect(
      screen.queryByRole("menuitem", { name: "Convert to story" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("menuitem", { name: "Edit item" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: "Delete item" })).toBeVisible();
  });

  it("keeps conversion locked when the linked story is unavailable", () => {
    setQuery([
      makeItem({ completed: true, convertedAt: "2026-10-13T01:00:00Z" }),
    ]);
    render(<StoryChecklist story={story} />);
    expect(screen.getByText("an unavailable story")).toBeVisible();
    expect(screen.getByRole("checkbox")).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Edit Review design" }),
    ).not.toBeInTheDocument();
  });

  it.each(["guest", "deleted", "archived"])(
    "shows existing items without mutation controls for %s stories",
    (mode) => {
      if (mode === "guest") mockUserRole = "guest";
      setQuery([makeItem()]);
      render(
        <StoryChecklist
          story={{
            ...story,
            deletedAt: mode === "deleted" ? "2026-10-13T00:00:00Z" : null,
            archivedAt: mode === "archived" ? "2026-10-13T00:00:00Z" : null,
          }}
        />,
      );
      expect(screen.getByText("Review design")).toBeVisible();
      expect(screen.getByRole("checkbox")).toBeDisabled();
      expect(screen.getAllByRole("button")).toHaveLength(1);
      expect(
        screen.getByRole("button", { name: "Checklist (1)" }),
      ).toHaveAttribute("aria-expanded", "true");
    },
  );

  it("requires confirmation before removing the entire checklist", () => {
    setQuery([makeItem()]);
    const { rerender } = render(<StoryChecklist story={story} />);
    openMenu("Checklist actions");
    fireEvent.click(screen.getByRole("menuitem", { name: "Delete checklist" }));
    expect(mockMutate).not.toHaveBeenCalled();
    const dialog = screen.getByRole("dialog", { name: "Delete checklist?" });
    expect(
      within(dialog).getByText(
        /Any stories created from these items will be kept/,
      ),
    ).toBeVisible();
    fireEvent.click(
      within(dialog).getByRole("button", { name: "Delete checklist" }),
    );
    expect(mockMutate).toHaveBeenCalledWith(
      { type: "clear" },
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
    setQuery([]);
    jest.mocked(useChecklistMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as unknown as ReturnType<typeof useChecklistMutation>);
    rerender(<StoryChecklist story={story} />);
    expect(
      screen.getByRole("dialog", { name: "Delete checklist?" }),
    ).toBeVisible();
    expect(screen.getByRole("button", { name: "Deleting..." })).toBeDisabled();
  });

  it("offers retry instead of treating a failed load as an empty checklist", () => {
    setQuery([], { isError: true });
    render(<StoryChecklist story={story} />);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Could not load this checklist.",
    );
    expect(
      screen.queryByRole("button", { name: "Add checklist" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(mockRefetch).toHaveBeenCalledTimes(1);
  });

  it("prevents conflicting actions while a checklist request is pending", () => {
    setQuery([makeItem()]);
    jest.mocked(useChecklistMutation).mockReturnValue({
      mutate: mockMutate,
      isPending: true,
    } as unknown as ReturnType<typeof useChecklistMutation>);
    render(<StoryChecklist story={story} />);
    expect(screen.getByRole("checkbox")).toBeDisabled();
    for (const button of screen.getAllByRole("button")) {
      if (button.textContent.startsWith("Checklist (")) continue;
      expect(button).toBeDisabled();
    }
  });

  it("persists the complete new order after keyboard sorting", async () => {
    const rectSpy = jest
      .spyOn(HTMLElement.prototype, "getBoundingClientRect")
      .mockImplementation(function getChecklistItemRect(this: HTMLElement) {
        const index =
          this.tagName === "LI"
            ? Array.from(this.parentElement!.children).indexOf(this)
            : 0;
        return {
          x: 0,
          y: index * 44,
          top: index * 44,
          bottom: index * 44 + 44,
          left: 0,
          right: 400,
          width: 400,
          height: 44,
          toJSON: () => ({}),
        };
      });
    setQuery([
      makeItem(),
      makeItem({ id: "item-2", title: "Test checkout", position: 1 }),
    ]);
    render(<StoryChecklist story={story} />);
    const handle = screen.getByRole("button", {
      name: "Reorder Review design",
    });
    handle.focus();
    fireEvent.keyDown(handle, { key: " ", code: "Space" });
    await waitFor(() => {
      expect(handle).toHaveAttribute("aria-pressed", "true");
    });
    fireEvent.keyDown(document, { key: "ArrowDown", code: "ArrowDown" });
    fireEvent.keyDown(document, { key: " ", code: "Space" });
    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith({
        type: "reorder",
        itemIds: ["item-2", "item-1"],
      });
    });
    rectSpy.mockRestore();
  });
});
