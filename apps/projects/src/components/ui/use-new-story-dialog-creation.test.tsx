import type { Editor } from "@tiptap/core";
import { act, renderHook } from "@testing-library/react";
import { toast } from "sonner";
import { useNewStoryDialogCreation } from "./use-new-story-dialog-creation";

jest.mock("sonner", () => ({
  toast: { warning: jest.fn(), error: jest.fn() },
}));
jest.mock("@/lib/tiptap/rich-text-media", () => ({
  clearRichTextContent: jest.fn(),
  getPersistableRichTextContent: () => ({
    contentText: "Draft",
    contentHtml: "<p>Draft</p>",
  }),
}));
jest.mock("./new-story-dialog-figma", () => ({
  attachFigmaDesigns: jest.fn(),
}));
const setup = (
  assigneeId: string | null,
  title = "Offline task",
  mayaAssigneeId?: string,
) => {
  const mutateStory = jest.fn().mockResolvedValue({ id: "created" });
  const focus = jest.fn();
  const onDialogClose = jest.fn();
  const props = {
    createMore: false,
    currentTeamId: "team",
    editor: {} as Editor,
    figmaArtifacts: [],
    finalizeStagedMedia: jest.fn().mockResolvedValue(null),
    isMayaAssigneeLoading: true,
    mayaAssigneeId,
    linkFigmaStory: jest.fn(),
    mutateStory,
    onDialogClose,
    onFormReset: jest.fn(),
    onFigmaArtifactsReset: jest.fn(),
    onResetDeadlineSource: jest.fn(),
    resetStagedMedia: jest.fn(),
    setStoryTitle: jest.fn(),
    storyForm: { title: "", priority: "No Priority" as const, assigneeId },
    storyTerm: "task",
    titleEditor: {
      commands: { focus, setContent: jest.fn() },
      getText: () => title,
    },
  };
  const view = renderHook(() => useNewStoryDialogCreation(props));
  return { ...view, mutateStory, focus, onDialogClose };
};
beforeEach(() => {
  jest.clearAllMocks();
});
it.each([null, "teammate"])(
  "creates ordinary work for %s while the optional Maya lookup is still pending",
  async (assignee) => {
    const view = setup(assignee);
    await act(async () => {
      await view.result.current.handleCreateStory();
    });
    expect(view.mutateStory).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Offline task", assigneeId: assignee }),
    );
    expect(view.onDialogClose).toHaveBeenCalledTimes(1);
  },
);
it("still validates an empty title while Maya metadata is loading", async () => {
  const view = setup(null, "");
  await act(async () => {
    await view.result.current.handleCreateStory();
  });
  expect(view.focus).toHaveBeenCalledTimes(1);
  expect(toast.warning).toHaveBeenCalledWith("Validation Error", {
    description: "Title is required",
  });
  expect(view.mutateStory).not.toHaveBeenCalled();
});
it("retains the preparation guard when the selected assignee is Maya", async () => {
  const view = setup("maya", "Maya task", "maya");
  await act(async () => {
    await view.result.current.handleCreateStory();
  });
  expect(view.mutateStory).not.toHaveBeenCalled();
});
