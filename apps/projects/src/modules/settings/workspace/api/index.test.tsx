/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { render, screen } from "@testing-library/react";
import { ApiSettings } from "./index";

const mockUseUserRole = jest.fn();

jest.mock("@/hooks/role", () => ({
  useUserRole: () => mockUseUserRole(),
}));

jest.mock("./components/personal-access-tokens", () => ({
  PersonalAccessTokens: () => <div>Personal tokens content</div>,
}));
jest.mock("./components/service-accounts", () => ({
  ServiceAccounts: () => <div>Service accounts content</div>,
}));
jest.mock("./components/webhook-endpoints", () => ({
  WebhookEndpoints: () => <div>Webhooks content</div>,
}));

describe("Developer settings", () => {
  beforeEach(() => {
    mockUseUserRole.mockReturnValue({ userRole: "admin" });
  });

  it("presents the released API capabilities to workspace administrators", () => {
    render(<ApiSettings />);

    expect(screen.getByRole("heading", { name: "API" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Access tokens" })).toBeVisible();
    expect(screen.getByRole("tab", { name: "Service accounts" })).toBeVisible();
    expect(
      screen.queryByRole("tab", { name: "OAuth applications" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Webhooks" })).toBeVisible();
    expect(screen.getByText("Personal tokens content")).toBeVisible();
    expect(screen.getByText("Security defaults")).toBeVisible();
  });

  it("keeps personal access tokens available to non-admin members", () => {
    mockUseUserRole.mockReturnValue({ userRole: "member" });

    render(<ApiSettings />);

    expect(screen.getByRole("tab", { name: "Access tokens" })).toBeVisible();
    expect(
      screen.queryByRole("tab", { name: "Service accounts" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("tab", { name: "OAuth applications" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("tab", { name: "Webhooks" }),
    ).not.toBeInTheDocument();
  });
});
