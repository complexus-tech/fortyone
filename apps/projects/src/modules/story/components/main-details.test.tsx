import type { ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { MainDetails } from "./main-details";

let mockEditorsReady = false;

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
  useEditor: ({ content }: { content: string }) =>
    mockEditorsReady
      ? {
          content,
          isFocused: false,
          getText: () => content,
          getHTML: () => content,
          commands: { setContent: jest.fn() },
        }
      : null,
}));
jest.mock("@/hooks", () => ({
  useLocalStorage: () => [true, jest.fn()],
  useUserRole: () => ({ userRole: "member" }),
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
  getRichTextContentType: () => "html",
}));
jest.mock("@/lib/tiptap/rich-text-media", () => ({
  RICH_TEXT_MEDIA_ACCEPT: "image/*",
}));
jest.mock("@/lib/tiptap/rich-text-table-menu", () => ({
  RichTextTableMenu: () => null,
}));
jest.mock("../hooks/update-mutation", () => ({
  useUpdateStoryMutation: () => ({ mutate: jest.fn() }),
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
