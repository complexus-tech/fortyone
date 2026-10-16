/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen, within } from "@testing-library/react";
import { CommentReactions } from "./comment-reactions";

describe("comment reactions", () => {
  it("shows only the add control until someone reacts, with all choices in its picker", () => {
    const onReact = jest.fn();
    render(
      <CommentReactions disabled={false} onReact={onReact} reactions={[]} />,
    );
    expect(screen.getAllByRole("button")).toHaveLength(1);
    fireEvent.click(screen.getByRole("button", { name: "Add reaction" }));
    const picker = screen.getByRole("dialog", { name: "Choose a reaction" });
    expect(within(picker).getAllByRole("button")).toHaveLength(80);
    fireEvent.click(within(picker).getByRole("button", { name: "Thumbs up" }));
    expect(onReact).toHaveBeenCalledWith("👍", true);
  });
  it("shows one aggregate chip per emoji and toggles the viewer's reaction", () => {
    const onReact = jest.fn();
    render(
      <CommentReactions
        disabled={false}
        onReact={onReact}
        reactions={[
          { emoji: "👍", count: 3, reacted: true },
          { emoji: "❤️", count: 2, reacted: false },
        ]}
      />,
    );
    expect(
      screen.getAllByRole("button", { name: "Thumbs up, 3 reactions" }),
    ).toHaveLength(1);
    expect(
      screen.queryByRole("button", { name: "React with thumbs up" }),
    ).not.toBeInTheDocument();
    fireEvent.click(
      screen.getByRole("button", { name: "Thumbs up, 3 reactions" }),
    );
    expect(onReact).toHaveBeenCalledWith("👍", false);
    fireEvent.click(screen.getByRole("button", { name: "Heart, 2 reactions" }));
    expect(onReact).toHaveBeenCalledWith("❤️", true);
  });
  it("searches other reactions in the shared popover and closes after selection", () => {
    const onReact = jest.fn();
    render(
      <CommentReactions disabled={false} onReact={onReact} reactions={[]} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Add reaction" }));
    const picker = screen.getByRole("dialog", { name: "Choose a reaction" });
    fireEvent.change(
      within(picker).getByRole("textbox", { name: "Search reactions" }),
      { target: { value: "celebrate" } },
    );
    expect(
      within(picker).queryByRole("button", { name: "Heart" }),
    ).not.toBeInTheDocument();
    fireEvent.click(within(picker).getByRole("button", { name: "Celebrate" }));
    expect(onReact).toHaveBeenCalledWith("🎉", true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
