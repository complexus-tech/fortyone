/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { GoogleDriveContextChips } from "./google-drive-context-chips";

const files = [
  {
    mimeType: "application/vnd.google-apps.document",
    name: "Project brief",
    referenceId: "document-reference",
  },
  {
    mimeType: "application/vnd.google-apps.spreadsheet",
    name: "Launch plan",
    referenceId: "spreadsheet-reference",
  },
];

describe("GoogleDriveContextChips", () => {
  it("uses native Google icons and a lightly tinted bordered surface", () => {
    render(<GoogleDriveContextChips files={files} />);

    const docChip = screen.getByText("Project brief").closest("span[title]");
    const sheetChip = screen.getByText("Launch plan").closest("span[title]");

    expect(docChip).toHaveClass(
      "border",
      "border-border/70",
      "bg-surface-muted/10",
      "gap-1.5",
      "rounded-2xl",
      "px-1.5",
      "py-1.5",
      "text-[0.95rem]",
    );
    expect(
      docChip?.querySelector('svg[viewBox="0 0 192 192"] path[fill="#3186FF"]'),
    ).not.toBeNull();
    expect(
      sheetChip?.querySelector(
        'svg[viewBox="0 0 192 192"] path[fill="#009954"]',
      ),
    ).not.toBeNull();
  });

  it("removes the selected Google file context", () => {
    const onRemove = jest.fn();
    render(<GoogleDriveContextChips files={files} onRemove={onRemove} />);

    fireEvent.click(screen.getByRole("button", { name: "Remove Launch plan" }));

    expect(onRemove).toHaveBeenCalledWith("spreadsheet-reference");
  });
});
