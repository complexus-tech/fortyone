/* global describe, expect, it -- Provided by the projects test runner. */
import { render } from "@testing-library/react";
import { FileTypeIcon } from "./file-type-icon";

describe("file type icons", () => {
  it("keeps repeated Office gradients scoped to each icon instance", () => {
    const { container } = render(
      <>
        <FileTypeIcon kind="word" />
        <FileTypeIcon kind="word" />
        <FileTypeIcon kind="excel" />
        <FileTypeIcon kind="powerpoint" />
      </>,
    );
    const gradients = Array.from(container.querySelectorAll("linearGradient"));
    expect(gradients).toHaveLength(4);
    expect(new Set(gradients.map((gradient) => gradient.id)).size).toBe(4);
    for (const gradient of gradients) {
      const icon = gradient.closest("svg")!;
      expect(
        Array.from(icon.querySelectorAll("path")).some(
          (path) => path.getAttribute("fill") === `url(#${gradient.id})`,
        ),
      ).toBe(true);
    }
  });

  it("renders all file kinds locally without adding accessible decoration or image requests", () => {
    const kinds = [
      "pdf",
      "word",
      "excel",
      "powerpoint",
      "csv",
      "text",
      "image",
      "video",
      "audio",
      "archive",
      "code",
      "link",
      "file",
    ] as const;
    const { container } = render(
      <>
        {kinds.map((kind) => (
          <FileTypeIcon key={kind} kind={kind} />
        ))}
      </>,
    );
    const icons = Array.from(container.querySelectorAll("svg"));
    expect(icons).toHaveLength(kinds.length);
    for (const icon of icons) {
      expect(icon.closest('[aria-hidden="true"]')).not.toBeNull();
      expect(icon).toHaveAttribute("focusable", "false");
      expect(icon.querySelector("path")).not.toBeNull();
    }
    expect(
      container.querySelector("img, image, use, script, foreignObject"),
    ).toBeNull();
  });
});
