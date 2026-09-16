/* global describe, expect, it -- Jest globals. */
import {
  assertChatAdmission,
  resolveChatMessageLimit,
  resolveChatTimezone,
} from "./request-context-policy";

describe("server-owned Maya context and admission policy", () => {
  it("validates device timezones and keeps UTC as the legacy default", () => {
    expect(resolveChatTimezone(undefined)).toBe("UTC");
    expect(resolveChatTimezone("Africa/Harare")).toBe("Africa/Harare");
    expect(() => resolveChatTimezone("Pretend/Admin")).toThrow("valid IANA");
    expect(() => resolveChatTimezone({})).toThrow("valid IANA");
  });
  it("selects existing effective plan rules from persisted billing and workspace trial", () => {
    const workspace = { trialEndsOn: "2026-10-01T00:00:00Z" };
    const now = Date.parse("2026-09-16T00:00:00Z");
    expect(resolveChatMessageLimit(workspace, null, now)).toBe(25);
    expect(
      resolveChatMessageLimit(
        workspace,
        { tier: "pro", status: "active" },
        now,
      ),
    ).toBe(100);
    expect(
      resolveChatMessageLimit(
        { trialEndsOn: null },
        { tier: "business", status: "canceled" },
        now,
      ),
    ).toBe(15);
    expect(
      resolveChatMessageLimit(
        workspace,
        { tier: "enterprise", status: "active" },
        now,
      ),
    ).toBe(Infinity);
  });
  it("rejects exhausted and unavailable usage, retaining the explicit internal exemption", () => {
    expect(() => {
      assertChatAdmission({ current: 15, limit: 15, isInternal: false });
    }).toThrow("monthly");
    expect(() => {
      assertChatAdmission({ current: NaN, limit: 15, isInternal: false });
    }).toThrow("unavailable");
    expect(() => {
      assertChatAdmission({ current: 14, limit: 15, isInternal: false });
    }).not.toThrow();
    expect(() => {
      assertChatAdmission({ current: 1000, limit: 15, isInternal: true });
    }).not.toThrow();
  });
});
