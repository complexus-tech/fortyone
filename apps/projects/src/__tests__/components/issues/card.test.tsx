import { render, screen } from "@testing-library/react";
import type { Issue } from "@/types/issue";
import { IssueCard } from "@/components/ui";

describe("Rendering IssueCard component with issue data", () => {
  const issue: Issue = {
    id: 1,
    title: "Issue 1",
    description: "This is issue 1",
    status: "In Progress",
  };
  it("should render the title of the issue", () => {
    render(<IssueCard issue={issue} />);
    const title = screen.getByText("Issue 1");

    expect(title).toBeInTheDocument();
  });

  it("should NOT render the description of the issue", () => {
    render(<IssueCard issue={issue} />);
    const description = screen.queryByText("This is issue 1");

    expect(description).not.toBeInTheDocument();
  });

  it("should render the status of the issue", () => {
    render(<IssueCard issue={issue} />);
    const status = screen.getByText("In Progress");

    expect(status).toBeInTheDocument();
  });
});
