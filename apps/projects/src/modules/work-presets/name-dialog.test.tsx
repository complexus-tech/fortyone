import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { PresetNameDialog } from "./name-dialog";

jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));

describe.each([
  { title: "Save task template", name: "Customer deal", shared: true },
  { title: "Save view", name: "Marketing pipeline", shared: false },
])("$title", ({ title, name, shared }) => {
  it("prevents dismissal, draft edits and duplicate saves while pending", async () => {
    let finish!: () => void;
    const save = jest.fn(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const onOpenChange = jest.fn();
    render(
      <PresetNameDialog
        onOpenChange={onOpenChange}
        onSave={save}
        open
        shared={shared}
        title={title}
      />,
    );
    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: `  ${name}  ` },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    const dialog = screen.getByRole("dialog");
    expect(dialog).toHaveAttribute("aria-busy", "true");
    expect(screen.getByLabelText("Name")).toBeDisabled();
    if (shared) expect(screen.getByLabelText("Visibility")).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Close" }),
    ).not.toBeInTheDocument();
    const saving = screen.getByRole("button", { name: "Saving..." });
    expect(saving).toBeDisabled();
    fireEvent.click(saving);
    fireEvent.keyDown(dialog, { key: "Escape" });
    fireEvent.pointerDown(document.body, { button: 0, pointerType: "mouse" });
    expect(save).toHaveBeenCalledTimes(1);
    expect(save).toHaveBeenCalledWith(name, "personal");
    expect(onOpenChange).not.toHaveBeenCalled();

    await act(async () => {
      finish();
    });
    expect(onOpenChange).toHaveBeenCalledTimes(1);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("retains the name and restores retry and dismissal after failure", async () => {
    const save = jest.fn().mockRejectedValue(new Error("Connection lost"));
    const onOpenChange = jest.fn();
    render(
      <PresetNameDialog
        initialName={name}
        onOpenChange={onOpenChange}
        onSave={save}
        open
        shared={shared}
        title={title}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => {
      expect(screen.getByRole("dialog")).toHaveAttribute("aria-busy", "false");
      expect(toast.error).toHaveBeenCalledWith("Could not save", {
        description: "Connection lost",
      });
    });
    expect(screen.getByLabelText("Name")).toHaveValue(name);
    expect(screen.getByLabelText("Name")).toBeEnabled();
    if (shared) expect(screen.getByLabelText("Visibility")).toBeEnabled();
    expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
    expect(onOpenChange).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onOpenChange).toHaveBeenCalledTimes(1);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
