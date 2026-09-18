import { renderHook } from "@testing-library/react";
import { HocuspocusProvider } from "@hocuspocus/provider";
import { getPublicEnv } from "@/public-env";
import { createCollaborationSessionAction } from "./actions";
import { useDocumentCollaboration } from "./use-document-collaboration";

jest.mock("@hocuspocus/provider", () => ({
  HocuspocusProvider: jest.fn(),
  WebSocketStatus: { Connected: "connected" },
}));
jest.mock("@/public-env", () => ({ getPublicEnv: jest.fn() }));
jest.mock("./actions", () => ({ createCollaborationSessionAction: jest.fn() }));

describe("documents without the optional collaboration service", () => {
  beforeEach(() => jest.clearAllMocks());

  it.each(["", "   "])("does not connect when the URL is %p", (url) => {
    jest.mocked(getPublicEnv).mockReturnValue({
      NODE_ENV: "test",
      API_URL: "https://api.example.com",
      ADMIN_URL: "https://admin.example.com",
      COLLABORATION_URL: url,
    });
    const { result, rerender, unmount } = renderHook(() =>
      useDocumentCollaboration("document-id", "workspace", {
        id: "user-id",
        name: "Alex",
      }),
    );
    rerender();
    expect(result.current.configured).toBe(false);
    expect(result.current.connection).toBeNull();
    expect(result.current.peers).toEqual([]);
    expect(createCollaborationSessionAction).not.toHaveBeenCalled();
    expect(HocuspocusProvider).not.toHaveBeenCalled();
    unmount();
  });
});
