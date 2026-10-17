import type { MouseEvent, ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast, type Action } from "sonner";
import { documentKeys } from "@/shared/documents/keys";
import {
  archiveDocumentAction,
  restoreArchivedDocumentAction,
} from "./actions";
import { useArchiveDocument, useRestoreArchivedDocument } from "./hooks";
import type { WorkspaceDocument } from "./types";

jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("@/lib/auth/client", () => ({ useSession: jest.fn() }));
jest.mock("./actions", () => ({
  archiveDocumentAction: jest.fn(),
  restoreArchivedDocumentAction: jest.fn(),
}));
jest.mock("./queries", () => ({}));

const restored = {
  id: "document",
  title: "Project brief",
  publicEnabled: false,
  publicToken: null,
  archivedAt: null,
} as WorkspaceDocument;

const setup = () => {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: Infinity },
      mutations: { retry: false },
    },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return { client, wrapper };
};

beforeEach(() => {
  jest.clearAllMocks();
});

it("offers Undo only after a successful archive and restores through the API", async () => {
  const { client, wrapper } = setup();
  const activeKey = documentKeys.list("acme");
  const archivedKey = documentKeys.list("acme", "", "archived");
  client.setQueryData(activeKey, [restored]);
  client.setQueryData(archivedKey, []);
  jest.mocked(archiveDocumentAction).mockResolvedValue({ data: null });
  jest
    .mocked(restoreArchivedDocumentAction)
    .mockResolvedValue({ data: restored });
  const { result } = renderHook(useArchiveDocument, { wrapper });

  await act(async () => {
    await result.current.mutateAsync("document");
  });
  expect(archiveDocumentAction).toHaveBeenCalledWith("document", "acme");
  expect(client.getQueryState(activeKey)?.isInvalidated).toBe(true);
  expect(client.getQueryState(archivedKey)?.isInvalidated).toBe(true);
  const options = jest.mocked(toast.success).mock.calls[0][1];
  expect(options?.action).toEqual(expect.objectContaining({ label: "Undo" }));
  const undo = options?.action as Action;
  act(() => {
    undo.onClick({} as MouseEvent<HTMLButtonElement>);
  });
  await waitFor(() => {
    expect(
      client.getQueryData(documentKeys.detail("acme", "document")),
    ).toEqual(restored);
  });
  expect(restoreArchivedDocumentAction).toHaveBeenCalledWith(
    "document",
    "acme",
  );
  expect(toast.success).toHaveBeenLastCalledWith("Document restored", {
    description: expect.stringContaining("Public sharing stays off"),
  });
});

it("does not offer Undo or invalidate caches after an archive error", async () => {
  const { client, wrapper } = setup();
  const key = documentKeys.list("acme");
  client.setQueryData(key, [restored]);
  jest
    .mocked(archiveDocumentAction)
    .mockResolvedValue({ error: { message: "Not allowed" } });
  const { result } = renderHook(useArchiveDocument, { wrapper });
  await act(async () => {
    await expect(result.current.mutateAsync("document")).rejects.toThrow(
      "Not allowed",
    );
  });
  expect(toast.success).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenCalledWith("Could not archive the document");
  expect(client.getQueryState(key)?.isInvalidated).toBe(false);
});

it("keeps archived caches unchanged and reports failed restoration", async () => {
  const { client, wrapper } = setup();
  const key = documentKeys.list("acme", "", "archived");
  client.setQueryData(key, [restored]);
  jest
    .mocked(restoreArchivedDocumentAction)
    .mockResolvedValue({ error: { message: "Already restored" } });
  const { result } = renderHook(useRestoreArchivedDocument, { wrapper });
  await act(async () => {
    await expect(result.current.mutateAsync("document")).rejects.toThrow(
      "Already restored",
    );
  });
  expect(client.getQueryData(key)).toEqual([restored]);
  expect(toast.success).not.toHaveBeenCalled();
  expect(toast.error).toHaveBeenCalledWith("Could not restore the document");
});
