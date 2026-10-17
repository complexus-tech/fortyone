/* global afterEach, describe, expect, it -- Jest globals are provided by the projects test runner. */

import { Editor } from "@tiptap/core";
import StarterKit from "@tiptap/starter-kit";
import {
  CommentHighlights,
  resolveCommentHighlight,
  updateCommentDraftHighlight,
  updateCommentHighlights,
} from "./comment-highlights";

describe("document comment highlights", () => {
  let editor: Editor;

  afterEach(() => {
    editor.destroy();
  });

  const createEditor = () => {
    editor = new Editor({
      extensions: [StarterKit, CommentHighlights],
      content: "<p>Before target after</p>",
    });
    return editor;
  };
  const highlightedText = (id: string) =>
    editor.view.dom.querySelector(`[data-comment-thread-id="${id}"]`)
      ?.textContent;

  it("keeps mapped thread positions when a draft is moved or dismissed", () => {
    createEditor();
    updateCommentHighlights(editor, [{ id: "thread", from: 8, to: 14 }]);
    updateCommentDraftHighlight(editor, { id: "draft", from: 15, to: 20 });

    editor.commands.insertContentAt(1, "New ");
    expect(highlightedText("thread")).toBe("target");
    expect(highlightedText("draft")).toBe("after");

    // The page remaps its draft after a transaction, without refetching threads.
    updateCommentDraftHighlight(editor, { id: "draft", from: 19, to: 24 });
    expect(highlightedText("thread")).toBe("target");
    expect(highlightedText("draft")).toBe("after");

    updateCommentDraftHighlight(editor, null);
    expect(highlightedText("thread")).toBe("target");
    expect(highlightedText("draft")).toBeUndefined();

    editor.commands.insertContentAt(1, "More ");
    expect(highlightedText("thread")).toBe("target");
  });

  it("replaces persisted threads independently of the mapped draft", () => {
    createEditor();
    updateCommentHighlights(editor, [{ id: "thread", from: 8, to: 14 }]);
    updateCommentDraftHighlight(editor, { id: "draft", from: 15, to: 20 });
    editor.commands.insertContentAt(1, "New ");

    updateCommentHighlights(editor, [{ id: "other", from: 5, to: 11 }]);
    expect(highlightedText("thread")).toBeUndefined();
    expect(highlightedText("other")).toBe("Before");
    expect(highlightedText("draft")).toBe("after");

    updateCommentHighlights(editor, []);
    expect(highlightedText("other")).toBeUndefined();
    expect(highlightedText("draft")).toBe("after");
  });

  it("ignores a detached draft without disturbing existing threads", () => {
    createEditor();
    updateCommentHighlights(editor, [{ id: "thread", from: 8, to: 14 }]);
    updateCommentDraftHighlight(editor, { id: "draft", from: 15, to: 20 });
    updateCommentDraftHighlight(editor, { id: "draft", from: 15, to: 15 });

    expect(highlightedText("thread")).toBe("target");
    expect(highlightedText("draft")).toBeUndefined();
  });
});

describe("quoted comment anchors", () => {
  let editor: Editor;
  afterEach(() => {
    editor.destroy();
  });
  const highlighted = () =>
    Array.from(
      editor.view.dom.querySelectorAll('[data-comment-thread-id="thread"]'),
    )
      .map((node) => node.textContent)
      .join("");
  const start = () => {
    editor = new Editor({
      extensions: [StarterKit, CommentHighlights],
      content: "<h2>Overview</h2><p>The original quoted passage.</p>",
    });
    const from = editor.state.doc.textContent.indexOf("The original") + 3;
    return {
      id: "thread",
      from,
      to: from + "The original quoted passage.".length,
      quote: "The original quoted passage.",
    };
  };

  it("keeps the exact passage after preceding paragraphs and a stale metadata refresh", () => {
    const anchor = start();
    updateCommentHighlights(editor, [anchor]);
    expect(highlighted()).toBe(anchor.quote);
    editor.commands.insertContentAt(
      0,
      "<p>Owner: Alice</p><p>Objective: Ship onboarding</p>",
    );
    updateCommentHighlights(editor, [anchor]);
    expect(highlighted()).toBe(anchor.quote);
    const located = resolveCommentHighlight(editor, anchor)!;
    expect(located.from).toBeGreaterThan(anchor.from);
    expect(editor.state.doc.textBetween(located.from, located.to)).toBe(
      anchor.quote,
    );
  });

  it("reattaches an exact quote after reopening with stale saved positions", () => {
    const anchor = start();
    editor.commands.insertContentAt(0, "<p>New preceding paragraph</p>");
    // No mapped decoration exists yet, as when reloading a saved document.
    updateCommentHighlights(editor, [anchor]);
    expect(highlighted()).toBe(anchor.quote);
  });

  it("does not invent an anchor for removed or ambiguous quoted text", () => {
    const anchor = start();
    editor.commands.setContent("<p>Changed paragraph.</p>");
    updateCommentHighlights(editor, [anchor]);
    expect(highlighted()).toBe("");
    expect(resolveCommentHighlight(editor, anchor)).toBeNull();
    editor.commands.setContent(
      `<p>New paragraph</p><p>${anchor.quote}</p><p>${anchor.quote}</p>`,
    );
    updateCommentHighlights(editor, [anchor]);
    expect(highlighted()).toBe("");
  });

  it("preserves a mapped occurrence when an identical quote is inserted before it", () => {
    const anchor = start();
    updateCommentHighlights(editor, [anchor]);
    editor.commands.insertContentAt(0, `<p>${anchor.quote}</p>`);
    updateCommentHighlights(editor, [anchor]);
    const located = resolveCommentHighlight(editor, anchor)!;
    expect(located.from).toBeGreaterThan(anchor.quote.length);
    expect(highlighted()).toBe(anchor.quote);
    editor.commands.insertContentAt(located.from, "Changed ");
    expect(highlighted()).toBe(anchor.quote);
    editor.commands.insertContentAt(located.from + 10, "changed ");
    expect(highlighted()).toBe("");
  });
});
