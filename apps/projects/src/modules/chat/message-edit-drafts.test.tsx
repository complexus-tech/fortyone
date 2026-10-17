/* global describe, expect, it, jest -- Provided by the projects test runner. */

import { fireEvent, render, screen, within } from "@testing-library/react";
import type { PreviewDraft } from "./chat-rich-text";
import type { MessageDraftScope } from "./message-edit-drafts";
import {
  MessageEditDraftProvider,
  useMessageEditDraft,
} from "./message-edit-drafts";

const richDraft = (text: string): PreviewDraft => ({
  text,
  content: {
    type: "doc",
    content: [
      {
        type: "paragraph",
        content: [{ type: "text", text, marks: [{ type: "bold" }] }],
      },
    ],
  },
});

const DraftProbe = ({
  label,
  messageId = "message-1",
  scope = "conversation",
  onRender,
}: {
  label: string;
  messageId?: string;
  scope?: MessageDraftScope;
  onRender?: () => void;
}) => {
  const { draft, setDraft, restoreDraft, clearMessageDrafts } =
    useMessageEditDraft(messageId, scope);
  onRender?.();

  return (
    <div data-testid={label}>
      <button
        onClick={() => {
          setDraft(richDraft("Original message"));
        }}
        type="button"
      >
        Edit
      </button>
      {draft ? (
        <>
          <input
            aria-label="Draft"
            onChange={(event) => {
              setDraft(richDraft(event.target.value));
            }}
            value={draft.text}
          />
          <output data-testid={`${label}-content`}>
            {JSON.stringify(draft.content)}
          </output>
          <button
            onClick={() => {
              setDraft(null);
            }}
            type="button"
          >
            Cancel
          </button>
        </>
      ) : null}
      <button onClick={clearMessageDrafts} type="button">
        Finish message
      </button>
      <button
        onClick={() => {
          restoreDraft(richDraft("Failed edit"));
        }}
        type="button"
      >
        Restore failed edit
      </button>
    </div>
  );
};

describe("virtualized message edit drafts", () => {
  it("restores failed saves without overwriting a newer draft", () => {
    render(
      <MessageEditDraftProvider>
        <DraftProbe label="row" />
      </MessageEditDraftProvider>,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Restore failed edit" }),
    );
    expect(screen.getByRole("textbox")).toHaveValue("Failed edit");
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "Newer draft" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Restore failed edit" }),
    );
    expect(screen.getByRole("textbox")).toHaveValue("Newer draft");
  });
  it("restores unsaved text and formatting after the row unmounts", () => {
    const fixture = (visible: boolean) => (
      <MessageEditDraftProvider>
        {visible ? <DraftProbe label="row" /> : null}
      </MessageEditDraftProvider>
    );
    const { rerender } = render(fixture(true));
    fireEvent.click(screen.getByRole("button", { name: "Edit" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Draft" }), {
      target: { value: "An unfinished revision" },
    });

    rerender(fixture(false));
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
    rerender(fixture(true));

    expect(screen.getByRole("textbox", { name: "Draft" })).toHaveValue(
      "An unfinished revision",
    );
    expect(screen.getByTestId("row-content")).toHaveTextContent(
      JSON.stringify(richDraft("An unfinished revision").content),
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    rerender(fixture(false));
    rerender(fixture(true));
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("notifies only the edited message and view while typing", () => {
    const otherMessageRender = jest.fn();
    const threadViewRender = jest.fn();
    render(
      <MessageEditDraftProvider>
        <DraftProbe label="editing" />
        <DraftProbe
          label="other-message"
          messageId="message-2"
          onRender={otherMessageRender}
        />
        <DraftProbe label="thread" onRender={threadViewRender} scope="thread" />
      </MessageEditDraftProvider>,
    );
    otherMessageRender.mockClear();
    threadViewRender.mockClear();
    const editing = within(screen.getByTestId("editing"));
    fireEvent.click(editing.getByRole("button", { name: "Edit" }));
    fireEvent.change(editing.getByRole("textbox"), {
      target: { value: "Typing a draft" },
    });

    expect(otherMessageRender).not.toHaveBeenCalled();
    expect(threadViewRender).not.toHaveBeenCalled();
    expect(
      within(screen.getByTestId("thread")).queryByRole("textbox"),
    ).not.toBeInTheDocument();
  });

  it("cancels one view independently and clears all copies when the message is saved or deleted", () => {
    render(
      <MessageEditDraftProvider>
        <DraftProbe label="conversation" />
        <DraftProbe label="thread" scope="thread" />
        <DraftProbe label="search" scope="search" />
      </MessageEditDraftProvider>,
    );
    const conversation = within(screen.getByTestId("conversation"));
    const thread = within(screen.getByTestId("thread"));
    const search = within(screen.getByTestId("search"));
    fireEvent.click(conversation.getByRole("button", { name: "Edit" }));
    fireEvent.click(thread.getByRole("button", { name: "Edit" }));
    fireEvent.click(search.getByRole("button", { name: "Edit" }));
    fireEvent.click(conversation.getByRole("button", { name: "Cancel" }));

    expect(conversation.queryByRole("textbox")).not.toBeInTheDocument();
    expect(thread.getByRole("textbox")).toBeInTheDocument();
    expect(search.getByRole("textbox")).toBeInTheDocument();

    fireEvent.click(
      conversation.getByRole("button", { name: "Finish message" }),
    );
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });

  it("keeps chat instances isolated and discards drafts when their provider unmounts", () => {
    const fixture = (session: string) => (
      <>
        <MessageEditDraftProvider key={session}>
          <DraftProbe label="first-chat" />
        </MessageEditDraftProvider>
        <MessageEditDraftProvider>
          <DraftProbe label="second-chat" />
        </MessageEditDraftProvider>
      </>
    );
    const { rerender } = render(fixture("first-session"));
    fireEvent.click(
      within(screen.getByTestId("first-chat")).getByRole("button", {
        name: "Edit",
      }),
    );
    expect(
      within(screen.getByTestId("second-chat")).queryByRole("textbox"),
    ).not.toBeInTheDocument();

    rerender(fixture("next-session"));
    expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
  });
});
