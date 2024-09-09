import { render, screen } from "@testing-library/react";
import { StoryRow } from "@/components/ui";
import type { Story } from "@/modules/stories/types";

describe("Rendering StoryRow component with story data", () => {
  const story: Story = {
    id: 1,
    title: "Story 1",
    description: "This is story 1",
    status: "In Progress",
  };
  it("should render the title of the story", () => {
    render(<StoryRow story={story} />);
    const title = screen.getByText("Story 1");

    expect(title).toBeInTheDocument();
  });

  it("should NOT render the description of the story", () => {
    render(<StoryRow story={story} />);
    const description = screen.queryByText("This is story 1");

    expect(description).not.toBeInTheDocument();
  });
});
