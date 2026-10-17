import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useNativeLink } from "@/components/ui/link-preview/native-query";
import { useNativePreviewTerminology } from "@/components/ui/link-preview/native-terminology";
import type { NativeLinkTarget } from "@/components/ui/link-preview/native-target";
import type { DetailedStory } from "@/shared/story/types";
import { ChatStoryPanel } from "./story-panel";

jest.mock("@/components/ui/link-preview/native-query", () => ({
  useNativeLink: jest.fn(),
}));
jest.mock("@/components/ui/link-preview/native-terminology", () => ({
  useNativePreviewTerminology: jest.fn(),
}));
jest.mock("@/modules/story", () => ({
  StoryPage: ({
    storyId,
    onClose,
  }: {
    storyId: string;
    onClose?: () => void;
  }) => (
    <div>
      <header>
        <span>TM-276</span>
        <button autoFocus onClick={onClose} type="button">
          Close story
        </button>
      </header>
      Story editor {storyId}
    </div>
  ),
}));

const target: Extract<NativeLinkTarget, { kind: "story" }> = {
  kind: "story",
  href: "https://acme.fortyone.app/work/TM-276",
  workspaceSlug: "acme",
  id: "TM-276",
};
const story = {
  id: "resolved-story-id",
  teamCode: "TM",
  sequenceId: 276,
} as DetailedStory;

beforeEach(() => {
  jest.mocked(useNativePreviewTerminology).mockReturnValue({
    label: "Story",
    term: "story",
  });
});

afterEach(() => {
  cleanup();
  jest.clearAllMocks();
});

it("mounts the existing editor with the resolved story identity inside the panel", async () => {
  jest.mocked(useNativeLink).mockReturnValue({
    data: { kind: "story", value: story },
    loading: false,
    unavailable: false,
    retry: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
    refreshing: false,
    refreshedAt: undefined,
    refreshError: null,
  });
  const onClose = jest.fn();
  render(<ChatStoryPanel onClose={onClose} target={target} />);
  expect(
    await screen.findByText("Story editor resolved-story-id"),
  ).toBeVisible();
  expect(screen.getAllByText("TM-276")).toHaveLength(1);
  expect(
    screen
      .getByRole("complementary", { name: "Story details" })
      .querySelectorAll("header"),
  ).toHaveLength(1);
  expect(screen.getByRole("button", { name: "Close story" })).toHaveFocus();
  fireEvent.keyDown(screen.getByRole("button", { name: "Close story" }), {
    key: "Escape",
  });
  expect(onClose).toHaveBeenCalledTimes(1);
});

it("does not mount an editor or reveal cached story details when access is denied", () => {
  const retry = jest.fn();
  jest
    .mocked(useNativeLink)
    .mockReturnValue({
      data: null,
      loading: false,
      unavailable: true,
      retry,
      refresh: jest.fn().mockResolvedValue(undefined),
      refreshing: false,
      refreshedAt: undefined,
      refreshError: null,
    });
  render(<ChatStoryPanel onClose={jest.fn()} target={target} />);
  expect(screen.queryByText(/Story editor/)).not.toBeInTheDocument();
  expect(screen.getByRole("alert")).toHaveTextContent("no longer have access");
  fireEvent.click(screen.getByRole("button", { name: "Try again" }));
  expect(retry).toHaveBeenCalledTimes(1);
});

it("uses workspace terminology for the loading state and close control", () => {
  jest.mocked(useNativePreviewTerminology).mockReturnValue({
    label: "Task",
    term: "task",
  });
  jest.mocked(useNativeLink).mockReturnValue({
    data: null,
    loading: true,
    unavailable: false,
    retry: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
    refreshing: false,
    refreshedAt: undefined,
    refreshError: null,
  });
  const onClose = jest.fn();
  render(<ChatStoryPanel onClose={onClose} target={target} />);
  expect(
    screen.getByRole("complementary", { name: "Task details" }),
  ).toBeVisible();
  expect(screen.getByRole("status")).toHaveTextContent("Loading task…");
  fireEvent.click(screen.getByRole("button", { name: "Close task" }));
  expect(onClose).toHaveBeenCalledTimes(1);
});

it("uses workspace terminology when the item is unavailable", () => {
  jest.mocked(useNativePreviewTerminology).mockReturnValue({
    label: "Ticket",
    term: "ticket",
  });
  jest.mocked(useNativeLink).mockReturnValue({
    data: null,
    loading: false,
    unavailable: true,
    retry: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
    refreshing: false,
    refreshedAt: undefined,
    refreshError: null,
  });
  render(<ChatStoryPanel onClose={jest.fn()} target={target} />);
  expect(screen.getByRole("alert")).toHaveTextContent(
    "This ticket is unavailable or you no longer have access.",
  );
  expect(screen.getByRole("button", { name: "Close ticket" })).toBeVisible();
});
