/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { WorkspaceImportSettings } from "./index";

jest.mock("next/image", () => ({
  __esModule: true,
  default: ({ alt }: { alt: string }) => <span aria-label={alt} role="img" />,
}));

jest.mock("./components/import-wizard", () => ({
  ImportWizard: ({ open }: { open: boolean }) =>
    open ? <div role="dialog">Import wizard</div> : null,
}));

describe("WorkspaceImportSettings", () => {
  it("presents one universal export importer and opens its shared wizard", () => {
    render(<WorkspaceImportSettings />);

    const importCard = screen.getByRole("region", {
      name: "Import issues from Jira, ClickUp, or anywhere",
    });

    expect(importCard).toHaveClass(
      "border-border",
      "bg-surface",
      "rounded-2xl",
      "border-[0.5px]",
    );
    expect(importCard).not.toHaveClass("rounded-3xl");
    for (const source of ["Jira", "ClickUp", "monday.com", "Asana"]) {
      expect(screen.getByRole("img", { name: source })).toBeInTheDocument();
    }

    expect(
      screen.queryByText("About direct Jira connections"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("Import boundaries")).not.toBeInTheDocument();
    expect(
      screen.queryByText("AI maps every upload automatically"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("Review and edit before importing"),
    ).not.toBeInTheDocument();

    const importButton = screen.getByRole("button", { name: "Import" });
    expect(importButton).toHaveClass(
      "bg-background-inverse",
      "text-foreground-inverse",
    );

    fireEvent.click(importButton);

    expect(screen.getByRole("dialog")).toHaveTextContent("Import wizard");
  });
});
