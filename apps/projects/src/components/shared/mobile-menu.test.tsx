/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import type * as UiModule from "ui";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { MobileMenuButton } from "./mobile-menu";

jest.mock("./sidebar/header", () => ({
  Header: () => <div>Workspace header</div>,
}));

jest.mock("./sidebar/navigation", () => ({
  Navigation: () => <nav aria-label="Workspace navigation">Navigation</nav>,
}));

jest.mock("./sidebar/teams", () => ({
  Teams: () => <div>Teams</div>,
}));

jest.mock("./workspace-actions", () => ({
  WorkspaceActions: ({ variant }: { variant: string }) => {
    const { useState } = jest.requireActual<typeof ReactModule>("react");
    const { Dialog } = jest.requireActual<typeof UiModule>("ui");
    const [isOpen, setIsOpen] = useState(false);

    return (
      <div data-workspace-actions-variant={variant}>
        <span>7 days left in trial</span>
        <button
          onClick={() => {
            setIsOpen(true);
          }}
          type="button"
        >
          Invite people
        </button>
        <Dialog onOpenChange={setIsOpen} open={isOpen}>
          <Dialog.Content aria-describedby={undefined}>
            <Dialog.Title>Invite people</Dialog.Title>
          </Dialog.Content>
        </Dialog>
      </div>
    );
  },
}));

describe("MobileMenuButton", () => {
  it("makes workspace actions available between the header and navigation", () => {
    render(<MobileMenuButton />);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));

    const menu = screen.getByRole("dialog", { name: "Mobile Menu" });
    const header = within(menu).getByText("Workspace header");
    const invite = within(menu).getByRole("button", { name: "Invite people" });
    const navigation = within(menu).getByRole("navigation", {
      name: "Workspace navigation",
    });

    expect(invite.closest("[data-workspace-actions-variant]")).toHaveAttribute(
      "data-workspace-actions-variant",
      "mobile",
    );
    expect(within(menu).getByText("7 days left in trial")).toBeInTheDocument();
    expect(within(menu).getByText("Teams")).toBeInTheDocument();
    expect(header.compareDocumentPosition(invite)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
    expect(invite.compareDocumentPosition(navigation)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
  });

  it("returns to the open menu after dismissing the invitation dialog", () => {
    render(<MobileMenuButton />);

    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));
    fireEvent.click(screen.getByRole("button", { name: "Invite people" }));

    const invitation = screen.getByRole("dialog", { name: "Invite people" });
    fireEvent.keyDown(invitation, { key: "Escape" });

    expect(
      screen.queryByRole("dialog", { name: "Invite people" }),
    ).not.toBeInTheDocument();
    const menu = screen.getByRole("dialog", { name: "Mobile Menu" });
    expect(
      within(menu).getByRole("button", { name: "Invite people" }),
    ).toBeInTheDocument();

    fireEvent.keyDown(menu, { key: "Escape" });

    expect(
      screen.queryByRole("dialog", { name: "Mobile Menu" }),
    ).not.toBeInTheDocument();
  });
});
