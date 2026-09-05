/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import { getGroupedStories } from "@/modules/stories/public/queries";
import { useWorkAttention } from "./use-work-attention";

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-a" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace-a" }),
}));
jest.mock("@/modules/stories/public/queries", () => ({
  getGroupedStories: jest.fn(),
}));

const renderAttention = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
  return renderHook(() => useWorkAttention(), { wrapper });
};

describe("work attention counts", () => {
  it("uses full filtered totals rather than the limited story preview", async () => {
    jest.mocked(getGroupedStories).mockImplementation(
      async (_, params) =>
        ({
          groups: [
            { totalCount: params.deadlineAfter ? 7 : 12, stories: [{}] },
          ],
        }) as unknown as Awaited<ReturnType<typeof getGroupedStories>>,
    );
    const { result } = renderAttention();
    await waitFor(() => {
      expect(result.current.isPending).toBe(false);
    });
    expect(result.current.items).toEqual([
      { view: "overdue", count: 12 },
      { view: "today", count: 7 },
    ]);
  });
  it("does not display failed requests as zero tasks", async () => {
    jest.mocked(getGroupedStories).mockRejectedValue(new Error("Unavailable"));
    const { result } = renderAttention();
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });
    expect(result.current.items.every(({ count }) => count === undefined)).toBe(
      true,
    );
  });
});
