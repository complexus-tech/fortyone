/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type * as ReactModule from "react";
import type { Story } from "@/modules/stories/types";
import { StoriesList } from "./stories-list";

jest.mock("ui", () => ({ Box: "div" }));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({
    data: [{ id: "team-web", code: "WEB", name: "Web" }],
  }),
}));
jest.mock("./story-dialog", () => ({ StoryDialog: () => null }));
jest.mock("./story/row", () => {
  const { useState } = jest.requireActual<typeof ReactModule>("react");
  return {
    StoryRow: ({ story, teamCode }: { story: Story; teamCode?: string }) => {
      const [expanded, setExpanded] = useState(false);
      return (
        <div>
          <span>{`${teamCode}-${story.sequenceId}`}</span>
          <button
            aria-expanded={expanded}
            onClick={() => {
              setExpanded(true);
            }}
            type="button"
          >
            {story.title}
          </button>
        </div>
      );
    },
  };
});

describe("StoriesList", () => {
  it("resolves a story team code when the payload only includes teamId", () => {
    const story = {
      id: "story-9",
      sequenceId: 9,
      teamId: "team-web",
      title: "Resolve workspace search references",
    } as Story;

    render(<StoriesList stories={[story]} />);

    expect(screen.getByText("WEB-9")).toBeInTheDocument();
  });
});

it("keeps row state when a realtime update changes the ticket title", () => {
  const story = {
    id: "story-a",
    title: "Original ticket",
    teamId: "team-a",
  } as Story;
  const { rerender } = render(<StoriesList stories={[story]} />);
  fireEvent.click(screen.getByRole("button", { name: story.title }));

  rerender(<StoriesList stories={[{ ...story, title: "Renamed ticket" }]} />);

  expect(
    screen.getByRole("button", { name: "Renamed ticket" }),
  ).toHaveAttribute("aria-expanded", "true");
});
