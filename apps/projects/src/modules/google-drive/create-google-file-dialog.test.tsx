/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { CreateGoogleFileDialog } from "./create-google-file-dialog";

const mockCreateFile = jest.fn();

jest.mock("./hooks", () => ({
  useCreateGoogleDriveFile: () => ({
    isPending: false,
    mutate: mockCreateFile,
  }),
}));

const target = { id: "story-1", type: "story" as const };

describe("CreateGoogleFileDialog", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    Object.defineProperty(globalThis.crypto, "randomUUID", {
      configurable: true,
      value: jest.fn(() => "00000000-0000-4000-8000-000000000001"),
    });
  });

  it("puts the file name first and uses the current Google file icons", () => {
    render(
      <CreateGoogleFileDialog
        initialFileType="document"
        isOpen
        onOpenChange={jest.fn()}
        target={target}
      />,
    );

    const dialog = screen.getByRole("dialog");
    const fileName = screen.getByRole("textbox", { name: "Google file name" });
    const fileType = screen.getByRole("group", { name: "File type" });
    const docOption = screen.getByRole("button", { name: /Google Doc/ });
    const sheetOption = screen.getByRole("button", { name: /Google Sheet/ });

    expect(dialog).toHaveClass("max-w-xl");
    expect(screen.getByText("Create Google file")).toBeInTheDocument();
    expect(Array.from(dialog.querySelectorAll("input, fieldset"))).toEqual([
      fileName,
      fileType,
    ]);
    expect(
      docOption.querySelector(
        'svg[viewBox="0 0 192 192"] path[fill="#3186FF"]',
      ),
    ).not.toBeNull();
    expect(
      sheetOption.querySelector(
        'svg[viewBox="0 0 192 192"] path[fill="#009954"]',
      ),
    ).not.toBeNull();
    expect(screen.getByText("Google Doc.")).toHaveClass("sr-only");
    expect(screen.getByText("Google Sheet.")).toHaveClass("sr-only");
  });

  it("creates the selected Google file with the entered name", () => {
    render(
      <CreateGoogleFileDialog
        initialFileType="document"
        isOpen
        onOpenChange={jest.fn()}
        target={target}
      />,
    );

    fireEvent.change(
      screen.getByRole("textbox", { name: "Google file name" }),
      { target: { value: "  Launch plan  " } },
    );
    fireEvent.click(screen.getByRole("button", { name: /Google Sheet/ }));
    fireEvent.click(screen.getByRole("button", { name: "Create Sheet" }));

    expect(mockCreateFile).toHaveBeenCalledWith(
      {
        fileType: "spreadsheet",
        idempotencyKey: "00000000-0000-4000-8000-000000000001",
        title: "Launch plan",
      },
      { onSuccess: expect.any(Function) },
    );
  });
});
