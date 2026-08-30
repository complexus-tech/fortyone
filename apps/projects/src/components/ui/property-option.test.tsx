/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { PropsWithChildren } from "react";
import { render, screen } from "@testing-library/react";
import { useMediaQuery } from "@/hooks/media";
import { PropertyOption } from "./property-option";

jest.mock("ui", () => ({
  Box: function MockBox({
    children,
    className,
  }: PropsWithChildren<{ className?: string }>) {
    return (
      <div className={className} data-testid="property-option-wrapper">
        {children}
      </div>
    );
  },
  Text: function MockText({
    children,
    className,
  }: PropsWithChildren<{ className?: string }>) {
    return <span className={className}>{children}</span>;
  },
}));

jest.mock("@/hooks/media", () => ({
  useMediaQuery: jest.fn(),
}));

const mockedUseMediaQuery = jest.mocked(useMediaQuery);

describe("PropertyOption", () => {
  beforeEach(() => {
    mockedUseMediaQuery.mockReturnValue(false);
  });

  it.each([
    { isCompact: true, isMobile: false, mode: "compact" },
    { isCompact: false, isMobile: true, mode: "mobile" },
  ])("renders only the value in $mode mode", ({ isCompact, isMobile }) => {
    mockedUseMediaQuery.mockReturnValue(isMobile);

    render(
      <PropertyOption
        isCompact={isCompact}
        isNotifications={false}
        label="Status"
        value={<button type="button">In progress</button>}
      />,
    );

    expect(screen.getByRole("button", { name: "In progress" })).toBeVisible();
    expect(screen.queryByText("Status")).not.toBeInTheDocument();
    expect(
      screen.queryByTestId("property-option-wrapper"),
    ).not.toBeInTheDocument();
  });

  it("keeps the desktop wrapper but suppresses its label for notifications", () => {
    render(
      <PropertyOption
        className="custom-option-class"
        isNotifications
        label="Status"
        value={<button type="button">In progress</button>}
      />,
    );

    expect(screen.getByRole("button", { name: "In progress" })).toBeVisible();
    expect(screen.queryByText("Status")).not.toBeInTheDocument();
    expect(screen.getByTestId("property-option-wrapper")).toHaveClass(
      "grid-cols-1",
      "custom-option-class",
    );
  });

  it("renders the label beside the value in the default desktop layout", () => {
    render(
      <PropertyOption
        isNotifications={false}
        label="Status"
        value={<button type="button">In progress</button>}
      />,
    );

    expect(screen.getByText("Status")).toBeVisible();
    expect(screen.getByRole("button", { name: "In progress" })).toBeVisible();
    expect(screen.getByTestId("property-option-wrapper")).not.toHaveClass(
      "grid-cols-1",
    );
  });
});
