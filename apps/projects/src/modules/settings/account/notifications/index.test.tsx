/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import type * as UI from "ui";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { NotificationsSettings } from "./index";

let mockTab = "email";
const mockSetTab = jest.fn();

jest.mock("nuqs", () => ({
  parseAsStringLiteral: () => ({
    withDefault: () => ({ withOptions: () => ({}) }),
  }),
  useQueryState: () => [mockTab, mockSetTab],
}));
jest.mock("icons", () => ({
  ChatIcon: () => null,
  EmailIcon: () => null,
  NotificationsUnreadIcon: () => null,
}));
jest.mock("@/modules/settings/components", () => ({
  SectionHeader: () => null,
}));
jest.mock("./components/email", () => ({
  EmailNotifications: () => <div>Email preferences</div>,
}));
jest.mock("ui", () => {
  const { Tabs } = jest.requireActual<typeof UI>("ui");
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return { Box: Container, Flex: Container, Text: Container, Tabs };
});

beforeEach(() => {
  mockTab = "email";
  mockSetTab.mockReset();
});
afterEach(cleanup);

describe("NotificationsSettings", () => {
  it("offers the Chat tab in the browser and updates the selected URL tab", () => {
    render(
      <NotificationsSettings chatNotifications={<div>Chat preferences</div>} />,
    );

    expect(screen.getByRole("tab", { name: "Email" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.queryByRole("tab", { name: "Desktop" })).toBeNull();
    fireEvent.mouseDown(screen.getByRole("tab", { name: "Chat" }), {
      button: 0,
      ctrlKey: false,
    });
    expect(mockSetTab).toHaveBeenCalledWith("chat");
  });

  it("opens Chat preferences when the URL selects the Chat tab", () => {
    mockTab = "chat";
    render(
      <NotificationsSettings chatNotifications={<div>Chat preferences</div>} />,
    );

    expect(screen.getByRole("tab", { name: "Chat" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("tabpanel")).toHaveTextContent("Chat preferences");
    expect(screen.queryByText("Email preferences")).toBeNull();
  });

  it("retains native Desktop preferences alongside Email and Chat", () => {
    mockTab = "desktop";
    render(
      <NotificationsSettings
        chatNotifications={<div>Chat preferences</div>}
        desktopNotifications={<div>Desktop preferences</div>}
      />,
    );

    expect(screen.getAllByRole("tab")).toHaveLength(3);
    expect(screen.getByRole("tabpanel")).toHaveTextContent(
      "Desktop preferences",
    );
  });

  it("shows Email when a desktop-only link is opened in the browser", () => {
    mockTab = "desktop";
    render(
      <NotificationsSettings chatNotifications={<div>Chat preferences</div>} />,
    );

    expect(screen.getByRole("tab", { name: "Email" })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("tabpanel")).toHaveTextContent("Email preferences");
  });
});
