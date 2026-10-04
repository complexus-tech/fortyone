/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type * as ReactModule from "react";
import type * as UiModule from "ui";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { ViewFavoritesProvider } from "@/shared/views/favorites-slot";
import { MobileMenuButton } from "./mobile-menu";

jest.mock("./sidebar/header", () => ({
  Header: () => <div>Workspace header</div>,
}));

jest.mock("./sidebar/navigation", () => ({
  Navigation: () => (
    <nav aria-label="Workspace navigation">
      <a
        href="/acme/my-work"
        onClick={(event) => {
          event.preventDefault();
        }}
      >
        My work
      </a>
    </nav>
  ),
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
            <a
              href="https://example.test/invitation-help"
              onClick={(event) => {
                event.preventDefault();
              }}
            >
              Invitation help
            </a>
          </Dialog.Content>
        </Dialog>
      </div>
    );
  },
}));

const Favorites = ({ isCollapsed }: { isCollapsed: boolean }) => (
  <section aria-label="Favorites" data-collapsed={isCollapsed}>
    <a href="/acme/teams/product/stories?view=board">
      <span>Product board</span>
    </a>
    <a
      href="/acme/teams/product/stories?view=active"
      onClick={(event) => {
        event.preventDefault();
      }}
    >
      Active view
    </a>
    <button type="button">Remove Product board from favorites</button>
  </section>
);

const renderWithFavorites = () =>
  render(
    <ViewFavoritesProvider Favorites={Favorites}>
      <MobileMenuButton />
    </ViewFavoritesProvider>,
  );

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

  it("composes expanded Favorites between navigation and teams inside the local sidebar", () => {
    renderWithFavorites();
    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));

    const menu = screen.getByRole("dialog", { name: "Mobile Menu" });
    const navigation = within(menu).getByRole("navigation");
    const favorites = within(menu).getByRole("region", { name: "Favorites" });
    const teams = within(menu).getByText("Teams");
    expect(favorites).toHaveAttribute("data-collapsed", "false");
    expect(favorites.closest("[data-sidebar-content]")).toContainElement(
      navigation,
    );
    expect(navigation.compareDocumentPosition(favorites)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
    expect(favorites.compareDocumentPosition(teams)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
  });

  it.each(["My work", "Product board", "Active view"])(
    "closes for ordinary navigation to %s, including prevented active-view selection",
    (name) => {
      renderWithFavorites();
      fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));
      const link = screen.getByRole("link", { name });
      fireEvent.click(link.firstElementChild ?? link);
      expect(
        screen.queryByRole("dialog", { name: "Mobile Menu" }),
      ).not.toBeInTheDocument();
    },
  );

  it.each([
    { metaKey: true },
    { ctrlKey: true },
    { shiftKey: true },
    { altKey: true },
    { button: 2 },
  ])("keeps the menu open for modified navigation %j", (modifiers) => {
    renderWithFavorites();
    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));
    fireEvent.click(
      screen.getByRole("link", { name: "Active view" }),
      modifiers,
    );
    expect(
      screen.getByRole("dialog", { name: "Mobile Menu" }),
    ).toBeInTheDocument();
  });

  it("keeps the menu open when removing a favorite", () => {
    renderWithFavorites();
    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));
    fireEvent.click(
      screen.getByRole("button", {
        name: "Remove Product board from favorites",
      }),
    );
    expect(
      screen.getByRole("dialog", { name: "Mobile Menu" }),
    ).toBeInTheDocument();
  });

  it("returns to the open menu after dismissing the invitation dialog", () => {
    render(<MobileMenuButton />);

    fireEvent.click(screen.getByRole("button", { name: "Mobile Menu" }));
    fireEvent.click(screen.getByRole("button", { name: "Invite people" }));

    const invitation = screen.getByRole("dialog", { name: "Invite people" });
    fireEvent.click(
      within(invitation).getByRole("link", { name: "Invitation help" }),
    );
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
