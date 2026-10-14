import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { toast } from "sonner";
import { HeaderActions } from "./header-actions";

const mockCopyText = jest.fn();
const mockOpenChat = jest.fn();
const mockDeleteObjective = jest.fn();
let mockUserRole = "admin";
let mockIsAdminOrOwner = true;
let mockDeletePending = false;

jest.mock("@/hooks", () => ({
  useCopyToClipboard: () => [null, mockCopyText],
  useTerminology: () => ({
    getTermDisplay: (_term: string, options?: { capitalize?: boolean }) =>
      options?.capitalize ? "Objective" : "objective",
  }),
  useUserRole: () => ({ userRole: mockUserRole }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
  }),
}));
jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: () => ({ isAdminOrOwner: mockIsAdminOrOwner }),
}));
jest.mock("@/context/chat-context", () => ({
  useChatActions: () => ({ openChat: mockOpenChat }),
}));
jest.mock("../hooks/delete-mutation", () => ({
  useDeleteObjectiveMutation: () => ({
    mutate: mockDeleteObjective,
    isPending: mockDeletePending,
  }),
}));
jest.mock("sonner", () => ({
  toast: { info: jest.fn(), error: jest.fn() },
}));

const objective = {
  id: "objective-1",
  name: "Launch the MVP",
  createdBy: "owner-1",
};

const openMenu = () => {
  fireEvent.keyDown(
    screen.getByRole("button", { name: "More objective actions" }),
    { key: "Enter" },
  );
};

beforeEach(() => {
  jest.clearAllMocks();
  mockCopyText.mockResolvedValue(true);
  mockUserRole = "admin";
  mockIsAdminOrOwner = true;
  mockDeletePending = false;
});

it("keeps Copy link and Delete together while requiring confirmation before deleting", async () => {
  render(<HeaderActions objective={objective} />);
  openMenu();

  expect(
    await screen.findByRole("menuitem", { name: "Copy link" }),
  ).toBeVisible();
  fireEvent.click(screen.getByRole("menuitem", { name: "Delete" }));

  expect(mockDeleteObjective).not.toHaveBeenCalled();
  expect(
    await screen.findByRole("dialog", { name: "Delete objective" }),
  ).toBeVisible();
  fireEvent.click(screen.getByRole("button", { name: "Yes, Delete" }));
  expect(mockDeleteObjective).toHaveBeenCalledWith("objective-1");
});

it("copies the current objective URL from the menu", async () => {
  render(<HeaderActions objective={objective} />);
  openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Copy link" }));

  await waitFor(() => {
    expect(mockCopyText).toHaveBeenCalledWith(window.location.href);
    expect(toast.info).toHaveBeenCalledWith("Success", {
      description: "Objective link copied to clipboard",
    });
  });
});

it("reports clipboard failure without reporting success", async () => {
  mockCopyText.mockResolvedValue(false);
  render(<HeaderActions objective={objective} />);
  openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Copy link" }));

  await waitFor(() => {
    expect(toast.error).toHaveBeenCalledWith("Unable to copy link", {
      description: "Please try again or copy the link from your address bar.",
    });
  });
  expect(toast.info).not.toHaveBeenCalled();
});

it("keeps navigation and copy available to guests while hiding Smart fill and unauthorized deletion", async () => {
  mockUserRole = "guest";
  mockIsAdminOrOwner = false;
  render(<HeaderActions objective={objective} />);

  expect(screen.getByRole("link", { name: "Map view" })).toHaveAttribute(
    "href",
    "/acme/strategy",
  );
  expect(
    screen.queryByRole("button", { name: "Smart fill" }),
  ).not.toBeInTheDocument();
  openMenu();
  expect(
    await screen.findByRole("menuitem", { name: "Copy link" }),
  ).toBeVisible();
  expect(
    screen.queryByRole("menuitem", { name: "Delete" }),
  ).not.toBeInTheDocument();
});

it("opens Smart fill with the objective context", () => {
  render(<HeaderActions objective={objective} />);
  fireEvent.click(screen.getByRole("button", { name: "Smart fill" }));
  expect(mockOpenChat).toHaveBeenCalledWith(
    'Suggest work items that can be added to the objective "Launch the MVP" from the team',
  );
});

it("disables repeat deletion while the mutation is pending", async () => {
  const { rerender } = render(<HeaderActions objective={objective} />);
  openMenu();
  fireEvent.click(await screen.findByRole("menuitem", { name: "Delete" }));
  mockDeletePending = true;
  rerender(<HeaderActions objective={objective} />);

  expect(screen.getByRole("button", { name: "Deleting..." })).toBeDisabled();
  expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
  fireEvent.click(screen.getByRole("button", { name: "Deleting..." }));
  expect(mockDeleteObjective).not.toHaveBeenCalled();
});
