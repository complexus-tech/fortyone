/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { ApiError } from "@/lib/http";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getWorkspaceAccess } from "./workspace-access";

jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
const ctx = {
  workspaceSlug: "acme",
  session: {},
  cookieHeader: "session=existing-account",
};

beforeEach(() => jest.clearAllMocks());

describe("workspace session access", () => {
  it("checks the scoped API with the existing account context before allowing workspace hydration", async () => {
    jest
      .mocked(getWorkspace)
      .mockResolvedValue({ id: "workspace" } as Awaited<
        ReturnType<typeof getWorkspace>
      >);
    await expect(getWorkspaceAccess(ctx)).resolves.toBe("allowed");
    expect(getWorkspace).toHaveBeenCalledWith(ctx);
  });

  it("provides an SSO recovery destination after scoped denial without granting tenant access", async () => {
    jest.mocked(getWorkspace).mockRejectedValue(
      new ApiError("Workspace SSO required", 403, {
        error: { code: "workspace_sso_required" },
      }),
    );
    await expect(getWorkspaceAccess(ctx)).resolves.toBe("sso-required");
    expect(ctx.cookieHeader).toBe("session=existing-account");
  });

  it("keeps other scoped session-policy denials separate from SSO recovery", async () => {
    jest
      .mocked(getWorkspace)
      .mockRejectedValue(new ApiError("Workspace unavailable", 404, null));
    await expect(getWorkspaceAccess(ctx)).resolves.toBe("denied");
  });

  it("surfaces API failures rather than misclassifying them as an SSO denial", async () => {
    const failure = new ApiError("API unavailable", 503, null);
    jest.mocked(getWorkspace).mockRejectedValue(failure);
    await expect(getWorkspaceAccess(ctx)).rejects.toBe(failure);
  });
});
