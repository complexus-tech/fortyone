import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { MessageReactions } from "./message-reactions";
import type { ChatPerson, ChatReaction } from "./types";

const person = (id: string, fullName: string): ChatPerson => ({
  id,
  fullName,
  avatarUrl: null,
  isSystem: false,
});
const reactions: ChatReaction[] = [
  {
    emoji: "👍",
    count: 2,
    reacted: true,
    users: [person("sarah", "Sarah Chen"), person("joseph", "Joseph Mukorivo")],
  },
  {
    emoji: "👀",
    count: 1,
    reacted: false,
    users: [person("alex", "Alex Morgan")],
  },
];

describe("MessageReactions", () => {
  it("shows only the names belonging to the hovered reaction", async () => {
    const onToggleReaction = jest.fn();
    render(
      <MessageReactions
        messageId="message"
        onToggleReaction={onToggleReaction}
        reactions={reactions}
      />,
    );
    const reaction = screen.getByRole("button", {
      name: "Remove 👍 reaction, 2 reactions",
    });
    fireEvent.pointerMove(reaction, { pointerType: "mouse" });
    const tooltip = await screen.findByRole("tooltip");
    expect(within(tooltip).getByText("Sarah Chen")).toBeInTheDocument();
    expect(within(tooltip).getByText("Joseph Mukorivo")).toBeInTheDocument();
    expect(within(tooltip).queryByText("Alex Morgan")).not.toBeInTheDocument();
    fireEvent.click(reaction);
    expect(onToggleReaction).toHaveBeenCalledWith("message", "👍");
  });

  it("shows names on keyboard focus and keeps legacy reactions usable", async () => {
    render(
      <MessageReactions
        messageId="message"
        onToggleReaction={jest.fn()}
        reactions={[
          reactions[1],
          { emoji: "🙌", count: 1, reacted: false },
          { emoji: "👍", count: 0, reacted: false, users: [] },
        ]}
      />,
    );
    act(() => {
      screen
        .getByRole("button", { name: "Add 👀 reaction, 1 reaction" })
        .focus();
    });
    expect(await screen.findByRole("tooltip")).toHaveTextContent("Alex Morgan");
    expect(
      screen.getByRole("button", { name: "Add 🙌 reaction, 1 reaction" }),
    ).toBeEnabled();
    expect(screen.getAllByRole("button")).toHaveLength(2);
  });
});
