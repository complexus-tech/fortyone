/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { substoryGenerationSchema } from ".";

describe("substoryGenerationSchema", () => {
  it("allows an intentional empty response but bounds generated substories", () => {
    const substory = { title: "Let users export their dashboard" };

    expect(substoryGenerationSchema.safeParse({ substories: [] }).success).toBe(
      true,
    );
    expect(
      substoryGenerationSchema.safeParse({
        substories: Array.from({ length: 6 }, () => substory),
      }).success,
    ).toBe(false);
    expect(
      substoryGenerationSchema.safeParse({
        substories: [{ title: "x".repeat(256) }],
      }).success,
    ).toBe(false);
    expect(
      substoryGenerationSchema.safeParse({ substories: [{ title: "   " }] })
        .success,
    ).toBe(false);
  });
});
