/** Public USD seat prices. Existing subscriptions retain their Stripe price. */
export const SUBSCRIPTION_PRICING = {
  pro: { monthly: 9, yearly: 86.4 },
  business: { monthly: 15, yearly: 144 },
} as const;
