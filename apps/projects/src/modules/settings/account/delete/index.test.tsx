import type { ButtonHTMLAttributes, ReactNode } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { getWorkspaces } from "@/lib/queries/get-workspaces";
import type { Workspace } from "@/types/workspace";
import { deleteAccount } from "./action";
import { DeleteAccountSettings } from "./index";

const mockRefresh = jest.fn();
const mockRouter = { refresh: mockRefresh };
jest.mock("next/navigation", () => ({ useRouter: () => mockRouter }));
jest.mock("@/lib/queries/get-workspaces", () => ({ getWorkspaces: jest.fn() }));
jest.mock("./action", () => ({ deleteAccount: jest.fn() }));
jest.mock("@/lib/hooks/profile", () => ({
  useProfile: () => ({
    data: { id: "stale-profile", email: "stale@example.com" },
  }),
}));
jest.mock("@/components/shared/sidebar/utils", () => ({
  clearAllStorage: jest.fn(),
}));
jest.mock("../../components", () => ({ SectionHeader: () => null }));
jest.mock("@/components/ui", () => ({
  ConfirmDialog: ({
    isOpen,
    onConfirm,
    errorMessage,
  }: {
    isOpen: boolean;
    onConfirm: () => void;
    errorMessage?: string;
  }) =>
    isOpen ? (
      <div>
        <button onClick={onConfirm} type="button">
          Confirm deletion
        </button>
        {errorMessage}
      </div>
    ) : null,
}));
jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  return {
    Box: Container,
    Text: Container,
    Button: ({
      children,
      href,
      loading,
      onClick,
      disabled,
    }: ButtonHTMLAttributes<HTMLButtonElement> & {
      href?: string;
      loading?: boolean;
    }) =>
      href ? (
        <a href={href}>{children}</a>
      ) : (
        <button disabled={disabled || loading} onClick={onClick} type="button">
          {children}
        </button>
      ),
  };
});

const account = { id: "verified-account", email: "verified@example.com" };
const workspace = (
  id: string,
  role: Workspace["userRole"] = "admin",
  deletedAt: string | null = null,
) =>
  ({
    id,
    name: `${id} workspace`,
    slug: id,
    userRole: role,
    deletedAt,
  }) as Workspace;
const renderSettings = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <DeleteAccountSettings account={account} mobileAuth />
    </QueryClientProvider>,
  );
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(getWorkspaces).mockResolvedValue([]);
});

it("shows the verified account and direct resolution links for active admin workspaces", async () => {
  jest
    .mocked(getWorkspaces)
    .mockResolvedValue([
      workspace("acme"),
      workspace("member", "member"),
      workspace("removed", "admin", "2026-09-01"),
    ]);
  renderSettings();
  expect(screen.getByText("verified@example.com")).toBeInTheDocument();
  expect(screen.queryByText("stale@example.com")).not.toBeInTheDocument();
  const transfer = await screen.findByRole("link", {
    name: "Manage administrators",
  });
  expect(transfer).toHaveAttribute(
    "href",
    "/acme/settings/workspace/members?callbackUrl=%2Fauth%2Faccount-deletion",
  );
  expect(
    screen.getByRole("link", { name: "Delete workspace" }),
  ).toHaveAttribute(
    "href",
    "/acme/settings?callbackUrl=%2Fauth%2Faccount-deletion#delete-workspace",
  );
  expect(screen.queryByText("member workspace")).not.toBeInTheDocument();
  expect(screen.queryByText("removed workspace")).not.toBeInTheDocument();
});

it("allows an account with no workspaces to reach deletion and retains actionable conflicts", async () => {
  jest
    .mocked(deleteAccount)
    .mockRejectedValue(
      new Error("Transfer administrator access for Acme first."),
    );
  renderSettings();
  await screen.findByText(/no active workspaces to administer/);
  fireEvent.click(screen.getByRole("button", { name: "Delete Account" }));
  fireEvent.click(screen.getByRole("button", { name: "Confirm deletion" }));
  await screen.findByText("Transfer administrator access for Acme first.");
  expect(deleteAccount).toHaveBeenCalledWith(account.id);
});

it("refreshes ownership after returning from workspace settings", async () => {
  jest
    .mocked(getWorkspaces)
    .mockResolvedValueOnce([workspace("acme")])
    .mockResolvedValue([]);
  renderSettings();
  await screen.findByText("acme workspace");
  fireEvent(window, new Event("pageshow"));
  await screen.findByText(/no active workspaces to administer/);
  expect(screen.queryByText("acme workspace")).not.toBeInTheDocument();
  expect(mockRefresh).toHaveBeenCalled();
});

it("surfaces loading failure and supports retry instead of claiming there are no workspaces", async () => {
  jest
    .mocked(getWorkspaces)
    .mockRejectedValueOnce(new Error("Unavailable"))
    .mockResolvedValue([]);
  renderSettings();
  await screen.findByText(/Could not load your workspaces/);
  expect(
    screen.queryByText(/no active workspaces to administer/),
  ).not.toBeInTheDocument();
  await waitFor(() => {
    expect(
      screen.getByRole("button", { name: "Refresh workspaces" }),
    ).toBeEnabled();
  });
  fireEvent.click(screen.getByRole("button", { name: "Refresh workspaces" }));
  await screen.findByText(/no active workspaces to administer/);
});
