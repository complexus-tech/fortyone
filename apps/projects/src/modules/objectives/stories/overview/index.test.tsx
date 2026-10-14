/* global afterAll, afterEach, beforeAll, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { Overview } from "./index";

type EditorOptions = {
  content: string;
  editable: boolean;
  onUpdate: (event: { editor: MockEditor }) => void;
  onBlur: () => void;
};

type MockEditor = {
  content: string;
  options: EditorOptions;
  getText: () => string;
  getHTML: () => string;
};

let mockIsDesktop = true;
let mockCanUpdate = true;
const mockMutate = jest.fn();
const mockObjective = {
  id: "objective-1",
  name: "Launch the marketplace",
  shortSummary: "Launch the first public release.",
  description: "<p>Build the catalog and checkout.</p>",
};

jest.mock("next/navigation", () => ({
  useParams: () => ({ objectiveId: "objective-1" }),
}));

jest.mock("@/hooks", () => ({
  useMediaQuery: () => mockIsDesktop,
  useTerminology: () => ({ getTermDisplay: () => "Objective" }),
}));

jest.mock("../../hooks", () => ({
  useCanUpdateObjective: () => mockCanUpdate,
  useObjective: () => ({ data: mockObjective }),
  useUpdateObjectiveMutation: () => ({ mutate: mockMutate }),
}));

jest.mock("@tiptap/react", () => ({
  useEditor: (options: EditorOptions) => {
    const { useRef } = jest.requireActual<{
      useRef: <T>(value: T) => { current: T };
    }>("react");
    const editorRef = useRef<MockEditor | null>(null);
    if (!editorRef.current) {
      editorRef.current = {
        content: options.content,
        options,
        getText() {
          return this.content;
        },
        getHTML() {
          return this.content;
        },
      };
    }
    editorRef.current.options = options;
    return editorRef.current;
  },
}));

jest.mock("@/lib/tiptap/starter-kit", () => ({
  createRichTextStarterKit: () => ({}),
}));

jest.mock("ui", () => ({
  Box: "div",
  Container: "main",
  Divider: "hr",
  TextEditor: ({
    asTitle,
    editor,
  }: {
    asTitle?: boolean;
    editor: MockEditor;
  }) => (
    <textarea
      aria-label={asTitle ? "Objective title" : "Objective description"}
      defaultValue={editor.content}
      onBlur={editor.options.onBlur}
      onChange={(event) => {
        editor.content = event.target.value;
        editor.options.onUpdate({ editor });
      }}
      readOnly={!editor.options.editable}
    />
  ),
}));

jest.mock("@/components/ui", () => ({
  BoardDividedPanel: Object.assign(
    ({ children }: { children: ReactNode }) => <div>{children}</div>,
    {
      MainPanel: ({ children }: { children: ReactNode }) => (
        <div>{children}</div>
      ),
      SideBar: ({ children }: { children: ReactNode }) => (
        <aside>{children}</aside>
      ),
    },
  ),
}));

jest.mock("../sidebar", () => ({ Sidebar: () => "Objective sidebar" }));
jest.mock("./activity", () => ({
  Activity: ({ viewport }: { viewport: string }) => (
    <section aria-label={`${viewport} activity`} />
  ),
}));
jest.mock("./properties", () => ({ Properties: () => null }));
jest.mock("@/modules/documents/related-documents", () => ({
  RelatedDocuments: () => null,
}));
jest.mock("@/modules/google-drive/public/files", () => ({
  GoogleDriveFileSection: () => null,
}));
jest.mock("@/modules/objectives/components/objective-forecast-risk", () => ({
  ObjectiveForecastRiskBanner: () => null,
}));

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);

beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: class {
      observe() {}
      disconnect() {}
    },
  });
});

afterAll(() => {
  if (resizeObserverDescriptor) {
    Object.defineProperty(
      globalThis,
      "ResizeObserver",
      resizeObserverDescriptor,
    );
  } else {
    Reflect.deleteProperty(globalThis, "ResizeObserver");
  }
});

const editAllFields = () => {
  fireEvent.change(screen.getByRole("textbox", { name: "Objective title" }), {
    target: { value: "Updated objective" },
  });
  fireEvent.change(
    screen.getByRole("textbox", { name: "Objective short summary" }),
    { target: { value: "Updated summary" } },
  );
  fireEvent.change(
    screen.getByRole("textbox", { name: "Objective description" }),
    { target: { value: "<p>Updated description</p>" } },
  );
};

const expectAllFieldsSaved = () => {
  expect(mockMutate).toHaveBeenCalledTimes(3);
  for (const data of [
    { name: "Updated objective" },
    { shortSummary: "Updated summary" },
    { description: "<p>Updated description</p>" },
  ]) {
    expect(mockMutate).toHaveBeenCalledWith({
      objectiveId: "objective-1",
      data,
    });
  }
};

describe("Objective overview", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    mockMutate.mockClear();
    mockIsDesktop = true;
    mockCanUpdate = true;
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it.each([true, false])(
    "mounts one title and description editor for the active viewport (desktop=%s)",
    (isDesktop) => {
      mockIsDesktop = isDesktop;
      render(<Overview />);

      expect(
        screen.getAllByRole("textbox", { name: "Objective title" }),
      ).toHaveLength(1);
      expect(
        screen.getAllByRole("textbox", { name: "Objective description" }),
      ).toHaveLength(1);
      expect(
        screen.getAllByRole("textbox", { name: "Objective short summary" }),
      ).toHaveLength(1);
      expect(
        screen.getByRole("region", {
          name: `${isDesktop ? "desktop" : "mobile"} activity`,
        }),
      ).toBeInTheDocument();
      expect(screen.queryByRole("complementary") !== null).toBe(isDesktop);
    },
  );

  it("saves rapid edits to different fields independently", () => {
    render(<Overview />);
    editAllFields();

    act(() => {
      jest.advanceTimersByTime(999);
    });
    expect(mockMutate).not.toHaveBeenCalled();
    act(() => {
      jest.advanceTimersByTime(1);
    });
    expectAllFieldsSaved();
  });

  it("flushes each field on blur without sending duplicate saves", () => {
    render(<Overview />);
    editAllFields();

    for (const name of [
      "Objective title",
      "Objective short summary",
      "Objective description",
    ]) {
      fireEvent.blur(screen.getByRole("textbox", { name }));
    }
    expectAllFieldsSaved();

    act(() => {
      jest.advanceTimersByTime(1000);
    });
    expectAllFieldsSaved();
  });

  it("flushes pending edits when leaving the overview tab", () => {
    const { rerender } = render(<Overview />);
    editAllFields();

    rerender(<section aria-label="Tickets" />);
    expectAllFieldsSaved();
    act(() => {
      jest.advanceTimersByTime(1000);
    });
    expectAllFieldsSaved();
  });

  it("retains drafts and pending saves when changing viewport", () => {
    const { rerender } = render(<Overview />);
    editAllFields();

    for (const isDesktop of [false, true]) {
      mockIsDesktop = isDesktop;
      rerender(<Overview />);
      expect(
        screen.getByRole("textbox", { name: "Objective title" }),
      ).toHaveValue("Updated objective");
      expect(
        screen.getByRole("textbox", { name: "Objective short summary" }),
      ).toHaveValue("Updated summary");
      expect(
        screen.getByRole("textbox", { name: "Objective description" }),
      ).toHaveValue("<p>Updated description</p>");
    }

    act(() => {
      jest.advanceTimersByTime(1000);
    });
    expectAllFieldsSaved();
  });

  it("keeps all fields read-only without update permission", () => {
    mockCanUpdate = false;
    render(<Overview />);

    for (const textbox of screen.getAllByRole("textbox")) {
      expect(textbox).toHaveAttribute("readonly");
    }
    expect(mockMutate).not.toHaveBeenCalled();
  });
});
