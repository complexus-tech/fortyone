/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { waitFor } from "@testing-library/react";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getCookieHeader } from "@/lib/http/header";
import { getWorkspaces } from "@/lib/queries/workspaces/get-workspaces";
import SettingsPage from "./page";

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http/header", () => ({
  getCookieHeader: jest.fn(),
}));

jest.mock("@/lib/queries/workspaces/get-workspaces", () => ({
  getWorkspaces: jest.fn(),
}));

jest.mock("@/modules/settings/workspace/general", () => ({
  WorkspaceGeneralSettings: () => <div>Workspace settings</div>,
}));

jest.mock("next/navigation", () => ({
  redirect: jest.fn(),
}));

const createDeferred = <T,>() => {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });

  return { promise, resolve };
};

const mockAuth = jest.mocked(auth);
const mockGetCookieHeader = jest.mocked(getCookieHeader);
const mockGetWorkspaces = jest.mocked(getWorkspaces);
const mockRedirect = jest.mocked(redirect);

describe("SettingsPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("resolves route and request context concurrently before loading workspaces", async () => {
    const params = createDeferred<{ workspaceSlug: string }>();
    const session = createDeferred<Awaited<ReturnType<typeof auth>>>();
    const cookieHeader = createDeferred<string>();
    mockAuth.mockReturnValue(session.promise);
    mockGetCookieHeader.mockReturnValue(cookieHeader.promise);
    mockGetWorkspaces.mockResolvedValue([
      {
        slug: "acme",
        userRole: "admin",
      },
    ] as unknown as Awaited<ReturnType<typeof getWorkspaces>>);

    const page = SettingsPage({
      params: params.promise,
    });

    await waitFor(() => {
      expect(mockAuth).toHaveBeenCalledTimes(1);
      expect(mockGetCookieHeader).toHaveBeenCalledTimes(1);
    });
    expect(mockGetWorkspaces).not.toHaveBeenCalled();

    params.resolve({ workspaceSlug: "acme" });
    session.resolve(null);
    cookieHeader.resolve("session=fortyone");

    await page;

    expect(mockGetWorkspaces).toHaveBeenCalledWith(
      undefined,
      "session=fortyone",
    );
    expect(mockRedirect).not.toHaveBeenCalled();
  });
});
