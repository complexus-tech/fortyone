/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Member } from "@/types";
import { useSession } from "@/lib/auth/client";
import { useMembers } from "@/lib/hooks/members";
import { documentKeys } from "@/shared/documents/keys";
import type { WorkspaceDocument } from "./types";
import { setDocumentPublicLinkAction } from "./actions";
import { useUpdateDocumentAccess } from "./hooks";
import { DocumentAccessMenu } from "./document-access-menu";
import { DocumentPublicLink } from "./document-public-link";
import { DocumentPageHeader } from "./document-page-header";

jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("@/lib/hooks/members", () => ({ useMembers: jest.fn() }));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("./hooks", () => ({ useUpdateDocumentAccess: jest.fn() }));
jest.mock("./actions", () => ({ setDocumentPublicLinkAction: jest.fn() }));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
// Use the shared controls themselves, excluding unrelated editor and chart exports.
jest.mock("ui", () => ({
  ...jest.requireActual("../../../../../packages/ui/src/avatar"),
  ...jest.requireActual("../../../../../packages/ui/src/box"),
  ...jest.requireActual("../../../../../packages/ui/src/button"),
  ...jest.requireActual("../../../../../packages/ui/src/flex"),
  ...jest.requireActual("../../../../../packages/ui/src/input"),
  ...jest.requireActual("../../../../../packages/ui/src/menu"),
  ...jest.requireActual("../../../../../packages/ui/src/popover"),
  ...jest.requireActual("../../../../../packages/ui/src/select"),
  ...jest.requireActual("../../../../../packages/ui/src/skeleton"),
  ...jest.requireActual("../../../../../packages/ui/src/switch"),
  ...jest.requireActual("../../../../../packages/ui/src/text"),
}));

const document: WorkspaceDocument = {
  id: "doc",
  workspaceId: "workspace",
  title: "Project brief",
  contentHtml: "<p>Launch plan</p>",
  contentText: "Launch plan",
  revision: 1,
  collaborationEpoch: 1,
  collaborative: true,
  publicToken: "public-token",
  visibility: "restricted",
  createdBy: "owner",
  updatedBy: "owner",
  createdAt: "2026-10-17T09:00:00Z",
  updatedAt: "2026-10-17T09:00:00Z",
  canEdit: true,
  sharedWith: [{ userId: "editor", role: "editor" }],
  relatedWork: [],
  relatedWorkCount: 0,
};
const member = (
  id: string,
  fullName: string,
  role: Member["role"] = "member",
): Member => ({
  id,
  fullName,
  role,
  username: fullName.toLowerCase(),
  email: `${id}@example.com`,
  avatarUrl: null,
  isActive: true,
  isSystem: false,
  isInternal: false,
  createdAt: "2026-10-17T09:00:00Z",
  updatedAt: "2026-10-17T09:00:00Z",
});
const mutate = jest.fn();
const members = [
  member("owner", "Joseph"),
  member("editor", "Daria"),
  member("guest", "Alex", "guest"),
];

const renderWithQuery = (children: React.ReactNode) => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const result = render(
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>,
  );
  return { ...result, queryClient };
};

beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useSession).mockReturnValue({
    data: { user: { id: "owner", name: "Joseph", image: null } },
  } as ReturnType<typeof useSession>);
  jest.mocked(useMembers).mockReturnValue({
    data: members,
    isPending: false,
    isError: false,
  } as ReturnType<typeof useMembers>);
  jest
    .mocked(useUpdateDocumentAccess)
    .mockReturnValue({ mutate, isPending: false } as unknown as ReturnType<
      typeof useUpdateDocumentAccess
    >);
  Object.defineProperties(HTMLElement.prototype, {
    hasPointerCapture: { configurable: true, value: () => false },
    setPointerCapture: { configurable: true, value: () => undefined },
    releasePointerCapture: { configurable: true, value: () => undefined },
    scrollIntoView: { configurable: true, value: () => undefined },
  });
});

describe("document sharing", () => {
  it("identifies the actual owner and keeps access unchanged until saved", async () => {
    jest.mocked(useSession).mockReturnValue({
      data: { user: { id: "editor", name: "Daria", image: null } },
    } as ReturnType<typeof useSession>);
    renderWithQuery(<DocumentAccessMenu document={document} />);
    fireEvent.click(screen.getByRole("button", { name: "Share document" }));
    expect(screen.getByText("Joseph")).toBeInTheDocument();
    expect(screen.queryByText("Daria (You)")).not.toBeInTheDocument();
    expect(
      screen.getByText(
        "Set whether invited members can comment or edit. Guests can only view.",
      ),
    ).toBeInTheDocument();
    fireEvent.keyDown(
      screen.getByRole("combobox", { name: "Access for Daria" }),
      { key: "Enter" },
    );
    fireEvent.click(await screen.findByRole("option", { name: "Can comment" }));
    expect(
      screen.getByRole("combobox", { name: "Access for Daria" }),
    ).toHaveTextContent("Can comment");
    expect(mutate).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    expect(mutate).toHaveBeenCalledWith(
      {
        visibility: "restricted",
        members: [{ userId: "editor", role: "viewer" }],
      },
      expect.any(Object),
    );
  });

  it("adds guests as viewers and restores the saved permissions on cancel", () => {
    renderWithQuery(<DocumentAccessMenu document={document} />);
    fireEvent.click(screen.getByRole("button", { name: "Share document" }));
    fireEvent.change(
      screen.getByRole("textbox", { name: "Find people to invite" }),
      { target: { value: "Alex" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Invite Alex" }));
    expect(
      screen.getByRole("combobox", { name: "Access for Alex" }),
    ).toBeDisabled();
    expect(
      screen.getByRole("combobox", { name: "Access for Alex" }),
    ).toHaveTextContent("Can view");
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    expect(mutate).toHaveBeenCalledWith(
      {
        visibility: "restricted",
        members: [
          { userId: "editor", role: "editor" },
          { userId: "guest", role: "viewer" },
        ],
      },
      expect.any(Object),
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(
      screen.queryByRole("combobox", { name: "Access for Alex" }),
    ).not.toBeInTheDocument();
  });

  it("exposes public link copying and preview actions", async () => {
    const writeText = jest.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      configurable: true,
      value: { writeText },
    });
    renderWithQuery(<DocumentPublicLink document={document} />);
    expect(
      screen.getByRole("link", { name: "Open preview in a new tab" }),
    ).toHaveAttribute("href", "/shared/docs/public-token");
    fireEvent.click(screen.getByRole("button", { name: "Copy public link" }));
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledWith(
        "http://localhost/shared/docs/public-token",
      );
    });
  });

  it("refreshes document lists when public access changes", async () => {
    const saved = { ...document, publicToken: null };
    jest.mocked(setDocumentPublicLinkAction).mockResolvedValue(saved);
    const { queryClient } = renderWithQuery(
      <DocumentPublicLink document={document} />,
    );
    const invalidate = jest.spyOn(queryClient, "invalidateQueries");
    fireEvent.click(screen.getByRole("switch", { name: "Enable public link" }));
    await waitFor(() => {
      expect(setDocumentPublicLinkAction).toHaveBeenCalledWith(
        "doc",
        "acme",
        false,
      );
    });
    await waitFor(() => {
      expect(
        queryClient.getQueryData(documentKeys.detail("acme", "doc")),
      ).toEqual(saved);
    });
    expect(invalidate).toHaveBeenCalledWith({
      queryKey: documentKeys.lists("acme"),
    });
  });

  it("keeps the saved public state and reports a failed update", async () => {
    jest
      .mocked(setDocumentPublicLinkAction)
      .mockRejectedValue(new Error("Connection failed"));
    const { queryClient } = renderWithQuery(
      <DocumentPublicLink document={document} />,
    );
    queryClient.setQueryData(documentKeys.detail("acme", "doc"), document);
    fireEvent.click(screen.getByRole("switch", { name: "Enable public link" }));
    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(
        "Could not update public sharing",
      );
    });
    expect(
      queryClient.getQueryData(documentKeys.detail("acme", "doc")),
    ).toEqual(document);
    expect(
      screen.getByRole("switch", { name: "Enable public link" }),
    ).toBeChecked();
    expect(
      screen.getByRole("button", { name: "Copy public link" }),
    ).toBeEnabled();
  });

  it("shows native save feedback and public access without exposing the public token", () => {
    const { rerender } = renderWithQuery(
      <DocumentPageHeader
        canDuplicate
        canManageDocument={false}
        commentCount={2}
        document={{ ...document, publicEnabled: true, publicToken: null }}
        isDuplicating={false}
        onArchive={jest.fn()}
        onBack={jest.fn()}
        onCopyLink={jest.fn()}
        onDelete={jest.fn()}
        onDuplicate={jest.fn()}
        onShowComments={jest.fn()}
        onShowHistory={jest.fn()}
        peers={[]}
        saveStatus="Saved"
        title="Project brief"
      />,
    );
    expect(screen.getAllByText("Saved").length).toBeGreaterThan(0);
    expect(
      screen.getByLabelText("Document access: Public"),
    ).toBeInTheDocument();
    rerender(
      <DocumentPageHeader
        canDuplicate
        canManageDocument={false}
        commentCount={2}
        document={{
          ...document,
          publicEnabled: false,
          publicToken: null,
          visibility: "private",
        }}
        isDuplicating={false}
        onArchive={jest.fn()}
        onBack={jest.fn()}
        onCopyLink={jest.fn()}
        onDelete={jest.fn()}
        onDuplicate={jest.fn()}
        onShowComments={jest.fn()}
        onShowHistory={jest.fn()}
        peers={[]}
        saveStatus="Not saved"
        title="Project brief"
      />,
    );
    expect(screen.getAllByText("Not saved").length).toBeGreaterThan(0);
    expect(screen.queryByText("Saved")).not.toBeInTheDocument();
    expect(
      screen.getByLabelText("Document access: Private"),
    ).toBeInTheDocument();
  });
});
