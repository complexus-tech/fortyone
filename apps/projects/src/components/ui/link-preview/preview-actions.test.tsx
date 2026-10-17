/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { toast } from "sonner";
import { PreviewActions } from "./preview-actions";

jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
afterEach(() => {
  cleanup();
  jest.clearAllMocks();
});

const openMenu = () => {
  fireEvent.keyDown(screen.getByRole("button", { name: "Preview options" }), {
    key: "Enter",
  });
};

describe("preview refresh actions", () => {
  it("refreshes without navigation and reports success only after completion", async () => {
    let resolve!: () => void;
    const refresh = jest.fn(
      () =>
        new Promise<void>((done) => {
          resolve = done;
        }),
    );
    render(
      <PreviewActions
        onRefresh={refresh}
        refreshedAt={Date.UTC(2026, 9, 17, 10)}
      />,
    );
    openMenu();
    expect(screen.getByText(/Last refreshed/)).toBeVisible();
    fireEvent.click(screen.getByRole("menuitem", { name: "Refresh preview" }));
    await waitFor(() => {
      expect(refresh).toHaveBeenCalledTimes(1);
    });
    expect(toast.success).not.toHaveBeenCalled();
    resolve();
    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith("Preview refreshed");
    });
  });

  it("shows feedback when refresh fails without claiming success", async () => {
    const refresh = jest.fn().mockRejectedValue(new Error("Unavailable"));
    render(<PreviewActions onRefresh={refresh} />);
    openMenu();
    fireEvent.click(screen.getByRole("menuitem", { name: "Refresh preview" }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Could not refresh preview. Please try again.",
      );
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("disables duplicate refresh while another card is refreshing the same preview", () => {
    const refresh = jest.fn();
    render(<PreviewActions onRefresh={refresh} refreshing />);
    fireEvent.keyDown(
      screen.getByRole("button", { name: "Refreshing preview" }),
      { key: "Enter" },
    );
    const item = screen.getByRole("menuitem", { name: "Refreshing preview…" });
    expect(item).toHaveAttribute("data-disabled");
    fireEvent.click(item);
    expect(refresh).not.toHaveBeenCalled();
  });
});
