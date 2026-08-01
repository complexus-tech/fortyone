/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { renderHook } from "@testing-library/react";
import { useUserRole } from "@/hooks";
import { useCanUpdateObjective } from "./use-can-update-objective";

jest.mock("@/hooks", () => ({
  useUserRole: jest.fn(),
}));

const mockedUseUserRole = jest.mocked(useUserRole);

describe("useCanUpdateObjective", () => {
  it.each(["admin", "member", "system"] as const)(
    "allows the %s role to update objectives",
    (userRole) => {
      mockedUseUserRole.mockReturnValue({ userRole });

      const { result } = renderHook(() => useCanUpdateObjective());

      expect(result.current).toBe(true);
    },
  );

  it("keeps guests read-only without disabling controls while the role resolves", () => {
    mockedUseUserRole.mockReturnValueOnce({ userRole: "guest" });
    const guestResult = renderHook(() => useCanUpdateObjective());
    expect(guestResult.result.current).toBe(false);

    mockedUseUserRole.mockReturnValueOnce({ userRole: undefined });
    const unresolvedResult = renderHook(() => useCanUpdateObjective());
    expect(unresolvedResult.result.current).toBe(true);
  });
});
