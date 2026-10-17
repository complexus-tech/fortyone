/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
/* eslint-disable @next/next/no-img-element -- Native image errors exercise the next/image fallback. */
import type { ImgHTMLAttributes } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { useLinkMetadata } from "@/lib/hooks/link-metadata";
import { ExternalLinkPreview, PreviewImage } from "./external-preview";

jest.mock("@/lib/hooks/link-metadata", () => ({ useLinkMetadata: jest.fn() }));
jest.mock("next/image", () => ({
  __esModule: true,
  default: ({ alt, src, onError }: ImgHTMLAttributes<HTMLImageElement>) => (
    <img alt={alt} onError={onError} src={src} />
  ),
}));

describe("shared external-link favicons", () => {
  it("shows an icon and a single centered domain when metadata is unavailable", () => {
    jest
      .mocked(useLinkMetadata)
      .mockReturnValue({ data: {} } as ReturnType<typeof useLinkMetadata>);
    render(<ExternalLinkPreview url="https://example.com/notes" />);
    expect(screen.getAllByText("example.com")).toHaveLength(1);
    expect(
      screen.getByRole("link", { name: "Open example.com" }),
    ).toHaveAttribute("href", "https://example.com/notes");
    expect(screen.getByRole("link").querySelector("svg")).toBeInTheDocument();
  });

  it("retains an explicit link title without fabricating missing OG content", () => {
    jest
      .mocked(useLinkMetadata)
      .mockReturnValue({ data: {} } as ReturnType<typeof useLinkMetadata>);
    render(
      <ExternalLinkPreview
        title="Design notes"
        url="https://example.com/notes"
      />,
    );
    expect(screen.getByText("Design notes")).toBeVisible();
    expect(screen.getByText("example.com")).toBeVisible();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });

  it("replaces a broken favicon with the shared link icon at the same size", () => {
    const { container } = render(
      <PreviewImage
        className="size-8"
        icon
        src="/api/metadata/image?url=broken"
      />,
    );
    const image = container.querySelector("img")!;
    fireEvent.error(image);
    expect(container.querySelector("img")).toBeNull();
    expect(container.querySelector("svg")).toBeInTheDocument();
    expect(container.firstChild).toHaveClass("size-8");
  });

  it("tries a new favicon after the previous source failed", () => {
    const { container, rerender } = render(
      <PreviewImage icon src="/api/metadata/image?url=first" />,
    );
    fireEvent.error(container.querySelector("img")!);
    rerender(<PreviewImage icon src="/api/metadata/image?url=second" />);
    expect(container.querySelector("img")).toHaveAttribute(
      "src",
      "/api/metadata/image?url=second",
    );
  });

  it("uses the same fallback when no favicon is available", () => {
    const { container } = render(<PreviewImage className="size-8" icon />);
    expect(container.querySelector("img")).toBeNull();
    expect(container.querySelector("svg")).toBeInTheDocument();
    expect(container.firstChild).toHaveClass("size-8");
  });
});
