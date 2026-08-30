/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  getApprovalFingerprint,
  getPreparedApprovalFingerprint,
  type MutationToolApproval,
} from "./approval-fingerprint";

jest.mock("server-only", () => ({}));

const createApproval = (
  overrides: Partial<MutationToolApproval> = {},
): MutationToolApproval => ({
  approved: true,
  input: { teamId: "team-1", title: "Ship launch" },
  toolCallId: "call-1",
  toolName: "createStory",
  ...overrides,
});

describe("mutation approval fingerprints", () => {
  it("canonicalizes nested object keys without changing array order", () => {
    const first = createApproval({
      input: {
        metadata: { priority: 1, status: "ready" },
        stories: [
          { estimate: 2, title: "First" },
          { estimate: 3, title: "Second" },
        ],
      },
    });
    const sameValue = createApproval({
      input: {
        stories: [
          { title: "First", estimate: 2 },
          { title: "Second", estimate: 3 },
        ],
        metadata: { status: "ready", priority: 1 },
      },
    });
    const reversedStories = createApproval({
      input: {
        metadata: { status: "ready", priority: 1 },
        stories: [
          { title: "Second", estimate: 3 },
          { title: "First", estimate: 2 },
        ],
      },
    });

    expect(getApprovalFingerprint(first)).toBe(
      getApprovalFingerprint(sameValue),
    );
    expect(getApprovalFingerprint(first)).not.toBe(
      getApprovalFingerprint(reversedStories),
    );
  });

  it("binds the approval decision and tool name into the identity", () => {
    const approval = createApproval();

    expect(getApprovalFingerprint(approval)).not.toBe(
      getApprovalFingerprint(createApproval({ approved: false })),
    );
    expect(getApprovalFingerprint(approval)).not.toBe(
      getApprovalFingerprint(createApproval({ toolName: "updateStory" })),
    );
  });

  it("fingerprints prepared input as an approved mutation", () => {
    const input = { teamId: "team-1", title: "Ship launch" };

    expect(
      getPreparedApprovalFingerprint(
        createApproval({ approved: false, input }),
        input,
      ),
    ).toBe(getApprovalFingerprint(createApproval({ input })));
  });
});
