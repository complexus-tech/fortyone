import type { Subscription } from "@/types";

export const formatSubscriptionPrice = (price: Subscription["price"]) => {
  // This catalog uses USD. Do not guess minor-unit conversion for other currencies.
  if (
    !price ||
    price.currency.toLowerCase() !== "usd" ||
    price.intervalCount < 1
  ) {
    return null;
  }
  const amount = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    maximumFractionDigits: 2,
  }).format(price.unitAmount / 100);
  const period =
    price.intervalCount === 1
      ? price.interval
      : `${price.intervalCount} ${price.interval}s`;
  return `${amount} per user/${period}`;
};
