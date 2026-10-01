import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { SecurityProvisioning } from "./scim-panel";

const mint = jest.fn();
const mintReset = jest.fn();
const revoke = jest.fn();
const retry = jest.fn();
const credential = {
  id: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  name: "Directory",
  prefix: "f41_scim_prefix",
  expiresAt: "2090-10-01T00:00:00Z",
  revokedAt: null,
};
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.com",
}));
jest.mock("./scim-api", () => ({
  useSCIMStatus: () => ({
    isPending: false,
    isError: false,
    isSuccess: true,
    data: {
      credentials: [credential],
      managedUsers: 3,
      pendingSeatSync: true,
      seatSyncError: "Sync failed",
    },
  }),
  useSCIMMutations: () => ({
    mint: { mutate: mint, reset: mintReset, isPending: false, error: null },
    revoke: { mutate: revoke, reset: jest.fn(), isPending: false, error: null },
    retry: { mutate: retry, isPending: false, error: null },
  }),
}));
beforeEach(() => {
  jest.clearAllMocks();
});
it("shows a newly minted token once and clears it when the dialog closes", async () => {
  mint.mockImplementation((input, callbacks) => {
    callbacks.onSuccess({ token: "private-scim-token", credential });
  });
  render(<SecurityProvisioning />);
  expect(
    screen.getByText("https://api.example.com/scim/v2/acme"),
  ).toBeInTheDocument();
  fireEvent.click(screen.getByRole("button", { name: "Create token" }));
  fireEvent.change(screen.getByLabelText("Token name"), {
    target: { value: "Company directory" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Create token" }));
  await waitFor(() => {
    expect(screen.getByLabelText("Bearer token")).toHaveValue(
      "private-scim-token",
    );
  });
  expect(mint.mock.calls[0][0]).toEqual({
    name: "Company directory",
    lifetimeDays: 90,
  });
  fireEvent.click(screen.getByRole("button", { name: "Done" }));
  fireEvent.click(screen.getByRole("button", { name: "Create token" }));
  expect(screen.queryByLabelText("Bearer token")).not.toBeInTheDocument();
  expect(screen.getByLabelText("Token name")).toHaveValue("");
});
it("revokes the selected credential and exposes a seat sync retry", () => {
  render(<SecurityProvisioning />);
  fireEvent.click(screen.getByRole("button", { name: "Retry seat sync" }));
  expect(retry).toHaveBeenCalledTimes(1);
  fireEvent.click(screen.getByRole("button", { name: "Revoke" }));
  fireEvent.click(screen.getByRole("button", { name: "Revoke token" }));
  expect(revoke.mock.calls[0][0]).toBe(credential.id);
});
