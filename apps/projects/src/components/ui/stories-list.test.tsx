/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import { render, screen } from "@testing-library/react";
import type { Story } from "@/modules/stories/types";
import { StoriesList } from "./stories-list";

jest.mock("ui", () => {
  const React = jest.requireActual<typeof ReactModule>("react");

  return {
    Box: ({ children }: { children?: ReactModule.ReactNode }) =>
      React.createElement("div", null, children),
  };
});
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({
    data: [{ id: "team-web", code: "WEB", name: "Web" }],
  }),
}));
jest.mock("./story/row", () => ({
  StoryRow: ({ story, teamCode }: { story: Story; teamCode?: string }) => (
    <div>{`${teamCode}-${story.sequenceId}`}</div>
  ),
}));
jest.mock("./story-dialog", () => ({ StoryDialog: () => null }));

const story = {
  id: "story-9",
  sequenceId: 9,
  teamId: "team-web",
  title: "Resolve workspace search references",
} as Story;

describe("StoriesList", () => {
  it("resolves a story team code when the payload only includes teamId", () => {
    render(<StoriesList stories={[story]} />);

    expect(screen.getByText("WEB-9")).toBeInTheDocument();
  });
});
