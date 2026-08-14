/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ComponentPropsWithoutRef, ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { SettingsBackButton } from "./settings-back-button";

jest.mock("icons", () => ({
  ChevronLeftIcon: (props: ComponentPropsWithoutRef<"svg">) => (
    <svg {...props} />
  ),
}));

jest.mock("ui", () => ({
  Button: ({
    asIcon,
    children,
    color,
    href,
    rounded,
    size,
    variant,
    ...props
  }: ComponentPropsWithoutRef<"a"> & {
    asIcon?: boolean;
    children: ReactNode;
    color?: string;
    href: string;
    rounded?: string;
    size?: string;
    variant?: string;
  }) => (
    <a
      {...props}
      data-as-icon={asIcon ? "true" : undefined}
      data-color={color}
      data-rounded={rounded}
      data-size={size}
      data-variant={variant}
      href={href}
    >
      {children}
    </a>
  ),
}));

describe("SettingsBackButton", () => {
  it("renders an accessible compact link with the standard settings treatment", () => {
    render(
      <SettingsBackButton
        href="/acme/settings/integrations"
        label="Back to integrations"
      />,
    );

    const backButton = screen.getByRole("link", {
      name: "Back to integrations",
    });

    expect(backButton).toHaveAttribute("href", "/acme/settings/integrations");
    expect(backButton).toHaveAttribute("data-as-icon", "true");
    expect(backButton).toHaveAttribute("data-color", "tertiary");
    expect(backButton).toHaveAttribute("data-size", "sm");
    expect(backButton).toHaveAttribute("data-variant", "naked");
    expect(backButton).not.toHaveAttribute("data-rounded");
    expect(backButton).toHaveClass("bg-state-hover", "shrink-0");
  });
});
