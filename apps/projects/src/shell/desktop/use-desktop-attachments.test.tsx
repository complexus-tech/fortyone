import { act, renderHook, waitFor } from "@testing-library/react";
import type { StoryAttachment } from "@/modules/story/types";
import { getDesktopBridge } from "./bridge";
import { useDesktopWorkspaceScope } from "./workspace-scope";
import { useDesktopAttachments } from "./use-desktop-attachments";

jest.mock("./bridge", () => ({ getDesktopBridge: jest.fn() }));
jest.mock("./workspace-scope", () => ({ useDesktopWorkspaceScope: jest.fn() }));
const read = jest.fn();
const createObjectURL = jest.fn().mockReturnValue("blob:local-preview");
const revokeObjectURL = jest.fn();
const scope = { accountId: "alice", workspaceId: "one", workspaceSlug: "acme" };
const attachment: StoryAttachment = {
  id: "file",
  filename: "notes.txt",
  mimeType: "text/plain",
  size: 5,
  url: "",
  uploadedBy: "alice",
  createdAt: "2026-10-15",
  pending: true,
};
beforeEach(() => {
  jest.clearAllMocks();
  jest.mocked(useDesktopWorkspaceScope).mockReturnValue(scope);
  jest
    .mocked(getDesktopBridge)
    .mockReturnValue({ readAttachment: read } as unknown as NonNullable<
      ReturnType<typeof getDesktopBridge>
    >);
  URL.createObjectURL = createObjectURL;
  URL.revokeObjectURL = revokeObjectURL;
  read.mockResolvedValue({
    data: btoa("hello"),
    mimeType: "text/plain",
    fileName: "notes.txt",
  });
});

it("restores a local preview without persisting blob URLs and releases it on unmount", async () => {
  const source = [attachment];
  const hook = renderHook(() => useDesktopAttachments("story", source));
  await waitFor(() => {
    expect(hook.result.current[0].url).toBe("blob:local-preview");
  });
  expect(source[0].url).toBe("");
  expect(read).toHaveBeenCalledWith({
    accountId: "alice",
    workspaceId: "one",
    storyId: "story",
    attachmentId: "file",
  });
  hook.unmount();
  expect(revokeObjectURL).toHaveBeenCalledWith("blob:local-preview");
});

it("ignores a late native read after switching accounts", async () => {
  let finish!: (value: unknown) => void;
  read.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        finish = resolve;
      }),
  );
  const hook = renderHook(() => useDesktopAttachments("story", [attachment]));
  jest
    .mocked(useDesktopWorkspaceScope)
    .mockReturnValue({ ...scope, accountId: "bob" });
  read.mockResolvedValue(null);
  hook.rerender();
  await act(async () => {
    finish({
      data: btoa("private"),
      mimeType: "text/plain",
      fileName: "private.txt",
    });
  });
  expect(createObjectURL).not.toHaveBeenCalled();
  expect(hook.result.current[0].url).toBe("");
  hook.unmount();
});

it("uses retained native bytes after an upload acknowledgement and leaves ordinary browser files alone", async () => {
  const source = [
    { ...attachment, pending: false, url: "https://example.test/file" },
  ];
  const hook = renderHook(() => useDesktopAttachments("story", source));
  await waitFor(() => {
    expect(hook.result.current[0].url).toBe("blob:local-preview");
  });
  hook.unmount();
  jest.mocked(getDesktopBridge).mockReturnValue(null);
  read.mockClear();
  const browser = renderHook(() => useDesktopAttachments("story", source));
  expect(browser.result.current).toEqual(source);
  expect(read).not.toHaveBeenCalled();
  browser.unmount();
});
