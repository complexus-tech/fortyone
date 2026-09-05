/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import type { Workspace } from "@/types/workspace";
import type * as WorkspaceRouting from "./workspace-url";

describe.each([
  ["fortyone.app", "https://acme.fortyone.app"],
  ["localhost", "/acme"],
])("workspace defaults on %s", (domain, base) => {
  const loadRouting = () => {
    const original = process.env.NEXT_PUBLIC_DOMAIN;
    let routing: typeof WorkspaceRouting;
    try {
      process.env.NEXT_PUBLIC_DOMAIN = domain;
      jest.isolateModules(() => {
        routing =
          jest.requireActual<typeof WorkspaceRouting>("./workspace-url");
      });
    } finally {
      if (original === undefined) delete process.env.NEXT_PUBLIC_DOMAIN;
      else process.env.NEXT_PUBLIC_DOMAIN = original;
    }
    return routing!;
  };
  const workspaces = [
    { id: "first", slug: "first" },
    { id: "active", slug: "acme" },
  ] as Workspace[];

  it("uses Maya for workspace links and the last-used workspace", () => {
    const { buildWorkspaceUrl, getRedirectUrl } = loadRouting();
    expect(buildWorkspaceUrl("acme")).toBe(`${base}/maya`);
    expect(getRedirectUrl(workspaces, [], "active")).toBe(`${base}/maya`);
    expect(getRedirectUrl([workspaces[1]])).toBe(`${base}/maya`);
  });

  it("preserves explicit destinations and callbacks", () => {
    const { buildWorkspaceUrl, getRedirectUrl } = loadRouting();
    expect(buildWorkspaceUrl("acme", "/my-work")).toBe(`${base}/my-work`);
    expect(
      getRedirectUrl(workspaces, [], "active", "/my-work?view=board"),
    ).toBe("/my-work?view=board");
  });
});
