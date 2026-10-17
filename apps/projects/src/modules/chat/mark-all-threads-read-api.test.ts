/* global expect, it, jest -- Jest globals are provided by the projects test runner. */

import { post } from "@/lib/http";
import { markAllThreadsRead } from "./api";

jest.mock("@/lib/http", () => ({ post: jest.fn() }));

it("requests one workspace-wide snapshot without automatic transport retries", async () => {
  jest.mocked(post).mockResolvedValue({ data: null });
  const scope = { workspaceSlug: "workspace" };

  await expect(markAllThreadsRead(scope)).resolves.toBeUndefined();
  expect(post).toHaveBeenCalledWith("chat/threads/read", {}, scope, {
    retry: 0,
  });
});
