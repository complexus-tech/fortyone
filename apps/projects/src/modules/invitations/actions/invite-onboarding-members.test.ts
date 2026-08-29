/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { inviteMembers } from "./invite";
import { inviteOnboardingMembers } from "./invite-onboarding-members";

jest.mock("./invite", () => ({
  inviteMembers: jest.fn(),
}));

const inviteMembersMock = jest.mocked(inviteMembers);

describe("inviteOnboardingMembers", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("preserves the onboarding admin role and selected teams", async () => {
    const response = { data: null };
    inviteMembersMock.mockResolvedValue(response);

    await expect(
      inviteOnboardingMembers(
        ["ada@example.com", "grace@example.com"],
        ["team-1", "team-2"],
        "complexus",
      ),
    ).resolves.toBe(response);

    expect(inviteMembersMock).toHaveBeenCalledWith(
      [
        {
          email: "ada@example.com",
          role: "admin",
          teamIds: ["team-1", "team-2"],
        },
        {
          email: "grace@example.com",
          role: "admin",
          teamIds: ["team-1", "team-2"],
        },
      ],
      "complexus",
    );
  });
});
