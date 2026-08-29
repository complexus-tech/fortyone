/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { invitationKeys } from "./keys";

describe("invitationKeys", () => {
  it("keeps personal and workspace invitation caches distinct", () => {
    expect(invitationKeys.mine).toEqual(["my-invitations"]);
    expect(invitationKeys.pending("complexus")).toEqual([
      "invitations",
      "pending",
      "complexus",
    ]);
  });
});
