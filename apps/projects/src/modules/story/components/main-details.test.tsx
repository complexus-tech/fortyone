/* global afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import type { JSONContent } from "@tiptap/core";
import { Editor, Node } from "@tiptap/core";
import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { act, fireEvent, render, screen } from "@testing-library/react";
import type * as RichTextMedia from "@/lib/tiptap/rich-text-media";
import { getPersistableRichTextContent } from "@/lib/tiptap/rich-text-media";
import { MainDetails } from "./main-details";

let mockEditorsReady = false;
let mockContentType = "html";
let mockUserRole = "member";
const mockUpdateStory = jest.fn();
type EditorCallbacks = {
  content: string;
  onCreate?: (event: { editor: Editor }) => void;
  onUpdate: (event: { editor: Editor }) => void;
  onBlur: () => void;
};
let mockDescriptionCallbacks: EditorCallbacks;
let mockTitleCallbacks: EditorCallbacks;

jest.mock("ui", () => ({
  Box: "div",
  Container: ({ children }: { children: ReactNode }) => <main>{children}</main>,
  Divider: "hr",
  TextEditor: ({
    asTitle,
    editor,
  }: {
    asTitle?: boolean;
    editor: { content: string } | null;
  }) => {
    if (!editor) return null;
    return asTitle ? <h1>{editor.content}</h1> : <p>{editor.content}</p>;
  },
}));
jest.mock("@tiptap/react", () => ({
  useEditor: (options: EditorCallbacks) => {
    if (options.onCreate) mockDescriptionCallbacks = options;
    else mockTitleCallbacks = options;
    const { content } = options;
    return mockEditorsReady
      ? {
          content,
          isFocused: false,
          getText: () => content,
          getHTML: () => content,
          commands: { setContent: jest.fn() },
        }
      : null;
  },
}));
jest.mock("@/hooks", () => ({
  useLocalStorage: () => [true, jest.fn()],
  useUserRole: () => ({ userRole: mockUserRole }),
}));
jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: () => ({ isAdminOrOwner: true }),
}));
jest.mock("@/components/shared", () => ({ BodyContainer: "div" }));
jest.mock("@/lib/hooks/links", () => ({
  useLinks: () => ({ data: [], isLoading: false }),
}));
jest.mock("@/lib/tiptap/rich-text-extensions", () => ({
  createRichTextExtensions: () => [],
}));
jest.mock("@/lib/tiptap/markdown", () => ({
  getRichTextContentType: () => mockContentType,
}));
jest.mock("@/lib/tiptap/rich-text-media", () => {
  const actual = jest.requireActual<typeof RichTextMedia>(
    "@/lib/tiptap/rich-text-media",
  );
  return {
    ...actual,
    getPersistableRichTextContent: jest.fn(
      actual.getPersistableRichTextContent,
    ),
  };
});
jest.mock("@/lib/tiptap/rich-text-table-menu", () => ({
  RichTextTableMenu: () => null,
}));
jest.mock("../hooks/update-mutation", () => ({
  useUpdateStoryMutation: () => ({ mutate: mockUpdateStory }),
}));
jest.mock("../hooks/use-story-description-media", () => ({
  useStoryDescriptionMedia: () => ({
    inputRef: { current: null },
    handleMediaFiles: jest.fn(),
    openMediaPicker: jest.fn(),
  }),
}));
jest.mock("../hooks/story", () => ({
  useStoryById: (id: string) => ({
    data: {
      id,
      title: `Title ${id}`,
      descriptionHTML: `Description ${id}`,
      description: `Description ${id}`,
      teamId: "team",
      reporterId: "viewer",
    },
  }),
}));
jest.mock("@/modules/documents/related-documents", () => ({
  RelatedDocuments: () => null,
}));
jest.mock("@/modules/google-drive/public/files", () => ({
  GoogleDriveFileSection: () => null,
}));
jest.mock("@/modules/google-drive/public/editor", () => ({
  useGoogleDriveDescriptionPaste: () => ({ onPaste: jest.fn(), picker: null }),
}));
jest.mock("./use-figma-description-paste", () => ({
  useFigmaDescriptionPaste: () => jest.fn(),
}));
jest.mock("./activities", () => ({ Activities: () => null }));
jest.mock("./associations", () => ({ Associations: () => null }));
jest.mock("./attachments", () => ({ Attachments: () => null }));
jest.mock("./links", () => ({ Links: () => null }));
jest.mock("./sub-stories", () => ({ SubStories: () => null }));
jest.mock("./story-banners", () => ({ StoryBanners: () => null }));
jest.mock("./links-skeleton", () => ({ LinksSkeleton: () => null }));
jest.mock("./figma-section", () => ({ FigmaSection: () => null }));
jest.mock("./options-header", () => ({ OptionsHeader: () => null }));
jest.mock("./options", () => ({ Options: () => null }));
jest.mock("./checklist", () => ({
  StoryChecklist: ({ story }: { story: { id: string } }) => (
    <section aria-label="Checklist">Checklist {story.id}</section>
  ),
}));

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
const notifyResizes: (() => void)[] = [];

beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      constructor(private readonly callback: () => void) {
        notifyResizes.push(callback);
      }
      observe() {}
      unobserve() {}
      disconnect() {
        const index = notifyResizes.indexOf(this.callback);
        if (index !== -1) notifyResizes.splice(index, 1);
      }
    },
  });
});

afterAll(() => {
  if (resizeObserverDescriptor)
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  else Reflect.deleteProperty(globalThis, "ResizeObserver");
});

afterEach(() => jest.restoreAllMocks());

it("keeps one description and checklist through editor loading, refreshes, and story navigation", () => {
  mockEditorsReady = false;
  notifyResizes.length = 0;
  const errors = jest
    .spyOn(console, "error")
    .mockImplementation(() => undefined);
  jest
    .spyOn(HTMLElement.prototype, "scrollHeight", "get")
    .mockImplementation(() => (mockEditorsReady ? 640 : 0));
  const { rerender } = render(
    <MainDetails isNotifications={false} storyId="first" />,
  );
  expect(screen.queryByText("Description first")).not.toBeInTheDocument();
  expect(screen.getAllByRole("region", { name: "Checklist" })).toHaveLength(1);

  mockEditorsReady = true;
  rerender(<MainDetails isNotifications={false} storyId="first" />);
  act(() => {
    notifyResizes.forEach((notify) => {
      notify();
    });
  });
  const initialChildCount = screen.getByRole("main").childElementCount;

  for (const storyId of ["first", "second", "second", "third"]) {
    rerender(<MainDetails isNotifications={false} storyId={storyId} />);
    act(() => {
      notifyResizes.forEach((notify) => {
        notify();
      });
    });
    expect(screen.getAllByText(`Description ${storyId}`)).toHaveLength(1);
    expect(screen.getAllByRole("region", { name: "Checklist" })).toHaveLength(
      1,
    );
    expect(screen.getByRole("region", { name: "Checklist" })).toHaveTextContent(
      `Checklist ${storyId}`,
    );
    expect(screen.getAllByRole("button", { name: "See more" })).toHaveLength(1);
    expect(screen.getByRole("main").childElementCount).toBe(initialChildCount);
    fireEvent.click(screen.getByRole("button", { name: "See more" }));
    expect(screen.getByRole("button", { name: "See less" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    fireEvent.click(screen.getByRole("button", { name: "See less" }));
  }

  expect(errors.mock.calls.flat().join(" ")).not.toMatch(
    /same key|unique "key"/i,
  );
});

describe("description autosave", () => {
  const editors: Editor[] = [];
  const media = Node.create({
    name: "image",
    group: "block",
    atom: true,
    addAttributes: () => ({
      attachmentId: { default: null },
      isUploading: { default: false },
      src: { default: null },
    }),
    renderHTML: ({ HTMLAttributes }) => ["img", HTMLAttributes],
  });
  const createEditor = (content: string | JSONContent) => {
    const editor = new Editor({
      extensions: [Document, Paragraph, Text, media],
      content,
    });
    editors.push(editor);
    return editor;
  };
  const editDescription = (editor: Editor, content: string | JSONContent) => {
    editor.commands.setContent(content, { emitUpdate: false });
    mockDescriptionCallbacks.onUpdate({ editor });
  };

  beforeEach(() => {
    jest.useFakeTimers();
    jest.clearAllMocks();
    mockEditorsReady = false;
    mockContentType = "html";
    mockUserRole = "member";
  });

  afterEach(() => {
    editors.splice(0).forEach((editor) => {
      if (!editor.isDestroyed) editor.destroy();
    });
    jest.useRealTimers();
  });

  it("serializes only the latest description after the debounce while keeping title saves independent", () => {
    render(<MainDetails isNotifications={false} storyId="first" />);
    const editor = createEditor("Original");
    const titleEditor = createEditor("Updated title");

    act(() => {
      editDescription(editor, "First draft");
      editDescription(editor, "Second draft");
      mockTitleCallbacks.onUpdate({ editor: titleEditor });
      editDescription(editor, "Final draft");
      jest.advanceTimersByTime(999);
    });

    expect(getPersistableRichTextContent).not.toHaveBeenCalled();
    expect(mockUpdateStory).not.toHaveBeenCalled();
    act(() => {
      jest.advanceTimersByTime(1);
    });

    expect(getPersistableRichTextContent).toHaveBeenCalledTimes(1);
    expect(mockUpdateStory).toHaveBeenCalledTimes(2);
    expect(mockUpdateStory).toHaveBeenCalledWith({
      storyId: "first",
      payload: { title: "Updated title" },
    });
    expect(mockUpdateStory).toHaveBeenCalledWith({
      storyId: "first",
      payload: {
        descriptionHTML: "<p>Final draft</p>",
        description: "Final draft",
        reconcileDescriptionMedia: true,
      },
    });
  });

  it("flushes on blur without a duplicate delayed save", () => {
    render(<MainDetails isNotifications={false} storyId="first" />);
    const editor = createEditor("Original");
    act(() => {
      editDescription(editor, "Blurred draft");
      mockDescriptionCallbacks.onBlur();
      jest.advanceTimersByTime(1000);
    });

    expect(getPersistableRichTextContent).toHaveBeenCalledTimes(1);
    expect(mockUpdateStory).toHaveBeenCalledTimes(1);
    expect(mockUpdateStory).toHaveBeenCalledWith({
      storyId: "first",
      payload: {
        descriptionHTML: "<p>Blurred draft</p>",
        description: "Blurred draft",
        reconcileDescriptionMedia: true,
      },
    });
  });

  it("flushes the queued document on unmount even if the editor was already destroyed", () => {
    const { unmount } = render(
      <MainDetails isNotifications={false} storyId="first" />,
    );
    const editor = createEditor("Original");
    act(() => {
      editDescription(editor, "Last unsaved draft");
      editor.destroy();
      unmount();
      jest.advanceTimersByTime(1000);
    });

    expect(getPersistableRichTextContent).toHaveBeenCalledTimes(1);
    expect(mockUpdateStory).toHaveBeenCalledTimes(1);
    expect(mockUpdateStory).toHaveBeenCalledWith({
      storyId: "first",
      payload: {
        descriptionHTML: "<p>Last unsaved draft</p>",
        description: "Last unsaved draft",
        reconcileDescriptionMedia: true,
      },
    });
  });

  it("keeps the original story and document when an editor is reused before the story switch flush", () => {
    const { rerender } = render(
      <MainDetails isNotifications={false} storyId="first" />,
    );
    const editor = createEditor("Original");
    act(() => {
      editDescription(editor, "First story draft");
      editor.commands.setContent("Second story content", {
        emitUpdate: false,
      });
    });
    rerender(<MainDetails isNotifications={false} storyId="second" />);

    expect(mockUpdateStory).toHaveBeenNthCalledWith(1, {
      storyId: "first",
      payload: {
        descriptionHTML: "<p>First story draft</p>",
        description: "First story draft",
        reconcileDescriptionMedia: true,
      },
    });
    act(() => {
      editDescription(editor, "Second story draft");
      jest.advanceTimersByTime(1000);
    });
    expect(getPersistableRichTextContent).toHaveBeenCalledTimes(2);
    expect(mockUpdateStory).toHaveBeenCalledTimes(2);
    expect(mockUpdateStory).toHaveBeenNthCalledWith(2, {
      storyId: "second",
      payload: {
        descriptionHTML: "<p>Second story draft</p>",
        description: "Second story draft",
        reconcileDescriptionMedia: true,
      },
    });
  });

  it("omits uploading media and only reconciles after the uploaded document is queued", () => {
    render(<MainDetails isNotifications={false} storyId="first" />);
    const editor = createEditor("Original");
    const documentWithMedia = (isUploading: boolean): JSONContent => ({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [{ type: "text", text: "Keep this draft" }],
        },
        {
          type: "image",
          attrs: {
            attachmentId: isUploading ? null : "attachment",
            isUploading,
            src: isUploading
              ? "blob:pending-upload"
              : "https://example.com/image.png",
          },
        },
      ],
    });

    act(() => {
      editDescription(editor, documentWithMedia(true));
      jest.advanceTimersByTime(1000);
    });
    expect(mockUpdateStory).toHaveBeenNthCalledWith(1, {
      storyId: "first",
      payload: {
        descriptionHTML: "<p>Keep this draft</p>",
        description: "Keep this draft",
        reconcileDescriptionMedia: false,
      },
    });

    act(() => {
      editDescription(editor, documentWithMedia(false));
      jest.advanceTimersByTime(1000);
    });
    expect(mockUpdateStory).toHaveBeenNthCalledWith(2, {
      storyId: "first",
      payload: {
        descriptionHTML: expect.stringContaining(
          'src="https://example.com/image.png"',
        ),
        description: editor.getText(),
        reconcileDescriptionMedia: true,
      },
    });
    expect(mockUpdateStory).toHaveBeenCalledTimes(2);
  });

  it("still migrates legacy markdown immediately and does not write for a guest", () => {
    mockContentType = "markdown";
    const { rerender } = render(
      <MainDetails isNotifications={false} storyId="first" />,
    );
    const editor = createEditor("Converted markdown");
    act(() => mockDescriptionCallbacks.onCreate?.({ editor }));
    expect(mockUpdateStory).toHaveBeenCalledWith({
      storyId: "first",
      payload: {
        descriptionHTML: "<p>Converted markdown</p>",
        description: "Converted markdown",
        reconcileDescriptionMedia: true,
      },
    });

    mockUserRole = "guest";
    rerender(<MainDetails isNotifications={false} storyId="second" />);
    act(() => mockDescriptionCallbacks.onCreate?.({ editor }));
    expect(mockUpdateStory).toHaveBeenCalledTimes(1);
  });
});
