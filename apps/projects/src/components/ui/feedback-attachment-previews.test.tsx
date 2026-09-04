/* global afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { StrictMode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { FeedbackAttachmentPreviews } from "./feedback-attachment-previews";

describe("FeedbackAttachmentPreviews", () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  it("keeps selected-image previews valid through Strict Mode cleanup", async () => {
    const createObjectURL = jest
      .fn()
      .mockReturnValueOnce("blob:discarded-preview")
      .mockReturnValue("blob:active-preview");
    const revokeObjectURL = jest.fn();
    Object.defineProperty(URL, "createObjectURL", {
      configurable: true,
      value: createObjectURL,
    });
    Object.defineProperty(URL, "revokeObjectURL", {
      configurable: true,
      value: revokeObjectURL,
    });
    const file = new File(["image"], "evidence.png", { type: "" });
    const onRemove = jest.fn();

    const { unmount } = render(
      <StrictMode>
        <FeedbackAttachmentPreviews
          files={[file]}
          layout="page"
          onRemove={onRemove}
        />
      </StrictMode>,
    );

    const preview = await screen.findByRole("img", { name: "evidence.png" });
    expect(preview).toHaveAttribute("src", "blob:active-preview");
    expect(preview.closest(".grid")).toHaveClass("lg:grid-cols-6");
    fireEvent.click(
      screen.getByRole("button", { name: "Remove evidence.png" }),
    );
    expect(onRemove).toHaveBeenCalledWith(file);

    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:discarded-preview");
    expect(revokeObjectURL).toHaveBeenCalledWith("blob:active-preview");
  });

  it("uses three columns in the widget", () => {
    const file = new File(["notes"], "notes.txt", { type: "text/plain" });
    render(
      <FeedbackAttachmentPreviews
        files={[file]}
        layout="widget"
        onRemove={jest.fn()}
      />,
    );

    expect(screen.getByText("notes.txt").closest(".grid")).toHaveClass(
      "grid-cols-3",
    );
  });
});
