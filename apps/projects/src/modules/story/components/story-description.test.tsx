import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { StoryDescription } from "./story-description";

const resizeObserverDescriptor = Object.getOwnPropertyDescriptor(
  globalThis,
  "ResizeObserver",
);
let contentHeight = 0;
const resizeObservers: TestResizeObserver[] = [];

class TestResizeObserver implements ResizeObserver {
  readonly targets = new Set<Element>();

  constructor(readonly callback: ResizeObserverCallback) {
    resizeObservers.push(this);
  }

  observe(target: Element) {
    this.targets.add(target);
  }

  unobserve(target: Element) {
    this.targets.delete(target);
  }

  disconnect() {
    this.targets.clear();
  }
}

beforeAll(() => {
  Object.defineProperty(globalThis, "ResizeObserver", {
    configurable: true,
    value: TestResizeObserver,
  });
});

beforeEach(() => {
  contentHeight = 0;
  resizeObservers.length = 0;
  jest
    .spyOn(HTMLElement.prototype, "scrollHeight", "get")
    .mockImplementation(() => contentHeight);
});

afterEach(() => {
  jest.restoreAllMocks();
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

const resizeContent = (height: number) => {
  act(() => {
    contentHeight = height;
    for (const observer of resizeObservers) {
      if (observer.targets.size > 0) observer.callback([], observer);
    }
  });
};

const EditableDescription = () => (
  <div
    aria-label="Story description"
    contentEditable
    role="textbox"
    suppressContentEditableWarning
    tabIndex={0}
  >
    A description that can be edited inline.
  </div>
);

it.each([0, 120, 240])(
  "does not offer expansion for content measuring %i pixels",
  (height) => {
    contentHeight = height;
    render(
      <StoryDescription>
        <p>A short description.</p>
      </StoryDescription>,
    );

    expect(screen.getByText("A short description.")).toBeVisible();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  },
);

it("expands and collapses long content without replacing the inline editor", async () => {
  contentHeight = 640;
  render(
    <StoryDescription>
      <EditableDescription />
    </StoryDescription>,
  );

  const editor = screen.getByRole("textbox", { name: "Story description" });
  const expand = await screen.findByRole("button", { name: "See more" });
  expect(expand).toHaveAttribute("type", "button");
  expect(expand).toHaveAttribute("aria-expanded", "false");
  const contentId = expand.getAttribute("aria-controls");
  expect(contentId).toBeTruthy();
  expect(document.getElementById(contentId!)).toContainElement(editor);

  fireEvent.click(expand);
  const collapse = screen.getByRole("button", { name: "See less" });
  expect(collapse).toHaveAttribute("aria-expanded", "true");
  expect(collapse).toHaveAttribute("aria-controls", contentId);
  expect(screen.getByRole("textbox")).toBe(editor);

  fireEvent.click(collapse);
  expect(screen.getByRole("button", { name: "See more" })).toHaveAttribute(
    "aria-expanded",
    "false",
  );
  expect(screen.getByRole("textbox")).toBe(editor);
  expect(editor).toHaveTextContent("A description that can be edited inline.");
});

it("updates the expansion control when hidden content appears, media grows, or the layout shrinks", async () => {
  render(
    <StoryDescription>
      <p>A description whose rendered size changes.</p>
    </StoryDescription>,
  );
  expect(screen.queryByRole("button")).not.toBeInTheDocument();

  resizeContent(700);
  expect(
    await screen.findByRole("button", { name: "See more" }),
  ).toHaveAttribute("aria-expanded", "false");

  resizeContent(200);
  await waitFor(() => {
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  resizeContent(900);
  expect(
    await screen.findByRole("button", { name: "See more" }),
  ).toHaveAttribute("aria-expanded", "false");
});

it("expands when editing starts and stays expanded after focus leaves the editor", async () => {
  contentHeight = 640;
  render(
    <>
      <StoryDescription>
        <EditableDescription />
      </StoryDescription>
      <button type="button">Outside the description</button>
    </>,
  );
  await screen.findByRole("button", { name: "See more" });
  const editor = screen.getByRole("textbox");

  act(() => {
    editor.focus();
  });
  expect(editor).toHaveFocus();
  expect(screen.getByRole("button", { name: "See less" })).toHaveAttribute(
    "aria-expanded",
    "true",
  );

  act(() => {
    screen.getByRole("button", { name: "Outside the description" }).focus();
  });
  expect(editor).not.toHaveFocus();
  expect(screen.getByRole("button", { name: "See less" })).toHaveAttribute(
    "aria-expanded",
    "true",
  );
});

it("moves focus out of the editor before collapsing when the browser does not focus a clicked button", async () => {
  contentHeight = 640;
  const onEditorBlur = jest.fn();
  render(
    <StoryDescription>
      <div onBlur={onEditorBlur}>
        <EditableDescription />
      </div>
    </StoryDescription>,
  );
  await screen.findByRole("button", { name: "See more" });
  const editor = screen.getByRole("textbox");
  act(() => {
    editor.focus();
  });
  editor.textContent = "An updated description that must keep its draft.";
  fireEvent.input(editor);
  const toggle = screen.getByRole("button", { name: "See less" });
  expect(editor).toHaveFocus();

  // fireEvent.click does not perform a browser's automatic pointer focus.
  // This reproduces clicking See less while editing in Safari on macOS.
  fireEvent.click(toggle);

  expect(toggle).toHaveFocus();
  expect(editor).not.toHaveFocus();
  expect(onEditorBlur).toHaveBeenCalledTimes(1);
  expect(toggle).toHaveAccessibleName("See more");
  expect(toggle).toHaveAttribute("aria-expanded", "false");
  expect(screen.getByRole("textbox")).toBe(editor);
  expect(editor).toHaveTextContent(
    "An updated description that must keep its draft.",
  );

  act(() => {
    editor.focus();
  });
  expect(editor).toHaveFocus();
  expect(toggle).toHaveAccessibleName("See less");
  expect(toggle).toHaveAttribute("aria-expanded", "true");
});

it("reveals focused links in read-only descriptions", async () => {
  contentHeight = 640;
  render(
    <StoryDescription>
      <p>A read-only description.</p>
      <a href="https://example.com/notes">Meeting notes</a>
    </StoryDescription>,
  );
  await screen.findByRole("button", { name: "See more" });
  const link = screen.getByRole("link", { name: "Meeting notes" });

  act(() => {
    link.focus();
  });
  expect(link).toHaveFocus();
  expect(screen.getByRole("button", { name: "See less" })).toHaveAttribute(
    "aria-expanded",
    "true",
  );
});

it("keeps the description collapsed when the expansion button receives keyboard focus", async () => {
  contentHeight = 640;
  render(
    <StoryDescription>
      <EditableDescription />
    </StoryDescription>,
  );
  const expand = await screen.findByRole("button", { name: "See more" });

  act(() => {
    expand.focus();
  });
  expect(expand).toHaveFocus();
  expect(expand).toHaveAccessibleName("See more");
  expect(expand).toHaveAttribute("aria-expanded", "false");
});

it("starts collapsed when switching to another story", async () => {
  contentHeight = 640;
  const { rerender } = render(
    <StoryDescription key="first-story">
      <p>The first story.</p>
    </StoryDescription>,
  );
  fireEvent.click(await screen.findByRole("button", { name: "See more" }));
  expect(screen.getByRole("button", { name: "See less" })).toHaveAttribute(
    "aria-expanded",
    "true",
  );

  rerender(
    <StoryDescription key="second-story">
      <p>The second story.</p>
    </StoryDescription>,
  );
  expect(screen.queryByText("The first story.")).not.toBeInTheDocument();
  expect(screen.getByText("The second story.")).toBeVisible();
  expect(
    await screen.findByRole("button", { name: "See more" }),
  ).toHaveAttribute("aria-expanded", "false");
});

it("toggles multiple story surfaces independently and associates each control with its own content", async () => {
  contentHeight = 640;
  render(
    <>
      <StoryDescription>
        <p>Mobile description.</p>
      </StoryDescription>
      <StoryDescription>
        <p>Desktop description.</p>
      </StoryDescription>
    </>,
  );
  const toggles = await screen.findAllByRole("button", { name: "See more" });
  const contentIds = toggles.map((toggle) =>
    toggle.getAttribute("aria-controls"),
  );
  expect(contentIds).toHaveLength(2);
  expect(new Set(contentIds).size).toBe(2);
  expect(document.getElementById(contentIds[0]!)).toHaveTextContent(
    "Mobile description.",
  );
  expect(document.getElementById(contentIds[1]!)).toHaveTextContent(
    "Desktop description.",
  );

  fireEvent.click(toggles[0]);
  expect(toggles[0]).toHaveAttribute("aria-expanded", "true");
  expect(toggles[1]).toHaveAttribute("aria-expanded", "false");

  fireEvent.click(toggles[1]);
  expect(toggles[0]).toHaveAttribute("aria-expanded", "true");
  expect(toggles[1]).toHaveAttribute("aria-expanded", "true");

  fireEvent.click(toggles[0]);
  expect(toggles[0]).toHaveAttribute("aria-expanded", "false");
  expect(toggles[1]).toHaveAttribute("aria-expanded", "true");
});
