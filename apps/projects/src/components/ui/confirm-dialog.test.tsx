/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { ConfirmDialog } from "./confirm-dialog";

const dialogProps = {
  isOpen: true,
  title: "Delete team",
  description: "Permanently delete this team and its data.",
  confirmText: "Delete team",
  confirmPhrase: "i understand",
};

describe("ConfirmDialog", () => {
  it("accepts case and surrounding whitespace only when normalization is enabled", () => {
    const onConfirm = jest.fn();
    const { rerender } = render(
      <ConfirmDialog {...dialogProps} onConfirm={onConfirm} />,
    );
    const confirmButton = screen.getByRole("button", { name: "Delete team" });

    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "  I Understand  " },
    });
    expect(confirmButton).toBeDisabled();

    rerender(
      <ConfirmDialog
        {...dialogProps}
        normalizeConfirmPhrase
        onConfirm={onConfirm}
      />,
    );
    expect(confirmButton).toBeEnabled();
    fireEvent.click(confirmButton);
    expect(onConfirm).toHaveBeenCalledTimes(1);

    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "i stand" },
    });
    expect(confirmButton).toBeDisabled();
  });

  it("blocks repeat confirmation, cancellation, escape, and outside dismissal while pending", () => {
    const onClose = jest.fn();
    const onCancel = jest.fn();
    const onConfirm = jest.fn();
    const { rerender } = render(
      <ConfirmDialog
        {...dialogProps}
        onCancel={onCancel}
        onClose={onClose}
        onConfirm={onConfirm}
      />,
    );
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "i understand" },
    });
    rerender(
      <ConfirmDialog
        {...dialogProps}
        isLoading
        loadingText="Deleting team..."
        onCancel={onCancel}
        onClose={onClose}
        onConfirm={onConfirm}
      />,
    );

    const dialog = screen.getByRole("dialog");
    const confirmButton = screen.getByRole("button", {
      name: "Deleting team...",
    });
    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    expect(dialog).toHaveAttribute("aria-busy", "true");
    expect(confirmButton).toBeDisabled();
    expect(cancelButton).toBeDisabled();
    expect(screen.getByRole("textbox")).toBeDisabled();
    expect(screen.queryByRole("button", { name: "Close" })).toBeNull();
    fireEvent.click(confirmButton);
    fireEvent.click(cancelButton);
    fireEvent.keyDown(dialog, { key: "Escape" });
    fireEvent.pointerDown(document.body);

    expect(onConfirm).not.toHaveBeenCalled();
    expect(onCancel).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("keeps an error visible and clears the phrase after cancellation", () => {
    const onClose = jest.fn();
    const { rerender } = render(
      <ConfirmDialog
        {...dialogProps}
        errorMessage="Could not delete the team. Try again."
        onClose={onClose}
        onConfirm={jest.fn()}
      />,
    );

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Could not delete the team. Try again.",
    );
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "i understand" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onClose).toHaveBeenCalledTimes(1);
    rerender(<ConfirmDialog {...dialogProps} onConfirm={jest.fn()} />);
    expect(screen.getByRole("textbox")).toHaveValue("");
    expect(screen.getByRole("button", { name: "Delete team" })).toBeDisabled();
  });
});
