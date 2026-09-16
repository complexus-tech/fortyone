import { formatSubscriptionPrice } from "./subscription-price";

describe("subscription price display", () => {
  it.each([
    [700, "month", "$7.00 per user/month"],
    [900, "month", "$9.00 per user/month"],
    [1000, "month", "$10.00 per user/month"],
    [1500, "month", "$15.00 per user/month"],
    [6720, "year", "$67.20 per user/year"],
    [8640, "year", "$86.40 per user/year"],
    [9600, "year", "$96.00 per user/year"],
    [14400, "year", "$144.00 per user/year"],
  ] as const)(
    "formats actual amount %s/%s",
    (unitAmount, interval, expected) => {
      expect(
        formatSubscriptionPrice({
          unitAmount,
          interval,
          intervalCount: 1,
          currency: "usd",
        }),
      ).toBe(expected);
    },
  );
  it("does not substitute a catalogue rate when pricing is unavailable", () => {
    expect(formatSubscriptionPrice(undefined)).toBeNull();
  });
});
