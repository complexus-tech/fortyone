/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { selectCollaborators } from "./collaborator-selection";

const member = (id: string, username = id) => ({
  avatarUrl: null,
  fullName: `Member ${id}`,
  id,
  username,
});

const collaborator = (id: string, username = id) => ({
  ...member(id, username),
  isActive: true,
  isSystem: false,
});

describe("selectCollaborators", () => {
  it("returns an empty selection when a legacy record has null collaborator ids", () => {
    expect(selectCollaborators(null, [], [])).toEqual([]);
  });

  it("preserves the selected ID order and excludes unknown members", () => {
    expect(
      selectCollaborators(
        ["second", "missing", "first"],
        [],
        [member("first"), member("second")],
      ),
    ).toEqual([member("second"), member("first")]);
  });

  it("prefers the story collaborator snapshot over the member list", () => {
    expect(
      selectCollaborators(
        ["selected"],
        [collaborator("selected", "story-username")],
        [member("selected", "member-username")],
      ),
    ).toEqual([collaborator("selected", "story-username")]);
  });
});
