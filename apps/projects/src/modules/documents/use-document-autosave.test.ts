import { act, renderHook } from "@testing-library/react";
import { useUpdateDocument } from "./hooks";
import { useDocumentAutosave } from "./use-document-autosave";

jest.mock("./hooks", () => ({ useUpdateDocument: jest.fn() }));
const mutateAsync = jest.fn();

beforeEach(() => {
  jest.useFakeTimers();
  mutateAsync.mockReset();
  jest
    .mocked(useUpdateDocument)
    .mockReturnValue({ mutateAsync } as unknown as ReturnType<
      typeof useUpdateDocument
    >);
});
afterEach(() => jest.useRealTimers());

test("queued title and body saves use the last committed revision", async () => {
  let finishFirst!: (value: { data: { revision: number } }) => void;
  mutateAsync
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          finishFirst = resolve;
        }),
    )
    .mockResolvedValueOnce({ data: { revision: 9 } });
  const { result } = renderHook(() => useDocumentAutosave("doc", true));
  result.current.initializeRevision(7);
  const title = result.current.persist({ title: "Updated" });
  const body = result.current.persist({
    contentHtml: "<p>Body</p>",
    contentText: "Body",
  });
  await act(async () => {
    await Promise.resolve();
  });
  expect(mutateAsync).toHaveBeenCalledTimes(1);
  expect(mutateAsync).toHaveBeenNthCalledWith(1, {
    title: "Updated",
    expectedRevision: 7,
  });
  await act(async () => {
    finishFirst({ data: { revision: 8 } });
    await Promise.all([title, body]);
  });
  expect(mutateAsync).toHaveBeenNthCalledWith(2, {
    contentHtml: "<p>Body</p>",
    contentText: "Body",
    expectedRevision: 8,
  });
});

test("a failed save stops queued writes until the page is reloaded", async () => {
  mutateAsync.mockRejectedValue(new Error("Conflict"));
  const { result } = renderHook(() => useDocumentAutosave("doc", true));
  const title = result.current.persist({ title: "Updated" });
  const body = result.current.persist({ contentText: "Body" });
  await act(async () => {
    await expect(title).rejects.toThrow("Conflict");
    await expect(body).rejects.toThrow("Reload before saving again");
  });
  expect(mutateAsync).toHaveBeenCalledTimes(1);
});

test("pending title and body drafts flush on unmount while collaborative saves stay disabled", async () => {
  mutateAsync.mockResolvedValue({ data: { revision: 8 } });
  const legacy = renderHook(() => useDocumentAutosave("doc", true));
  act(() => {
    legacy.result.current.initializeRevision(7);
    legacy.result.current.saveTitle("Draft title");
    legacy.result.current.saveContent({
      contentHtml: "<p>Draft</p>",
      contentText: "Draft",
    });
  });
  expect(mutateAsync).not.toHaveBeenCalled();
  await act(async () => {
    legacy.unmount();
  });
  expect(mutateAsync).toHaveBeenCalledTimes(2);
  const collaborative = renderHook(() => useDocumentAutosave("doc", false));
  await collaborative.result.current.persist({ title: "Must use Yjs" });
  expect(mutateAsync).toHaveBeenCalledTimes(2);
});
