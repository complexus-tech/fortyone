/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  canSendMayaMessage,
  shouldShowMayaMessageLimit,
} from "./message-limit";

describe("Maya message limits", () => {
  it("allows internal users to send after the workspace reaches its AI message limit", () => {
    expect(
      canSendMayaMessage({
        isInternalUser: true,
        limit: 15,
        totalMessages: 15,
      }),
    ).toBe(true);
  });

  it("blocks non-internal users after the workspace reaches its AI message limit", () => {
    expect(
      canSendMayaMessage({
        isInternalUser: false,
        limit: 15,
        totalMessages: 15,
      }),
    ).toBe(false);
  });

  it("only shows the limit warning to non-internal users", () => {
    expect(
      shouldShowMayaMessageLimit({
        isInternalUser: true,
        limit: 15,
        totalMessages: 30,
      }),
    ).toBe(false);

    expect(
      shouldShowMayaMessageLimit({
        isInternalUser: false,
        limit: 15,
        totalMessages: 30,
      }),
    ).toBe(true);
  });
});
