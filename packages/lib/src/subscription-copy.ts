import { SUBSCRIPTION_PRICING } from "./subscription-pricing";

/** Shared customer-facing billing answers for the website and documentation. */
export const BILLING_FAQS = [
  {
    question: "How much do the paid plans cost?",
    answer: `Professional is $${SUBSCRIPTION_PRICING.pro.monthly} per user/month and Business is $${SUBSCRIPTION_PRICING.business.monthly} per user/month when billed monthly. Enterprise pricing is available on request.`,
  },
  {
    question: "Is there a discount for annual billing?",
    answer: `Annual billing saves 20%. Professional costs $${SUBSCRIPTION_PRICING.pro.yearly.toFixed(2)} per user/year ($${(SUBSCRIPTION_PRICING.pro.yearly / 12).toFixed(2)} per month equivalent). Business costs $${SUBSCRIPTION_PRICING.business.yearly} per user/year ($${SUBSCRIPTION_PRICING.business.yearly / 12} per month equivalent). The full annual amount is billed upfront.`,
  },
  {
    question: "Does the price increase affect my existing subscription?",
    answer:
      "Existing subscriptions keep their current per-user rate while remaining on the same plan and billing interval. You can see your subscription's rate on the Billing page and manage payment details through Manage subscription.",
  },
  {
    question: "What happens when I add more members?",
    answer:
      "Additional billable members use your existing subscription's per-user rate, including any retained rate from before the price increase. Normal prorated charges apply for the remaining billing period. Your plan's member limit still applies; guests are not billable seats.",
  },
  {
    question: "What happens if I change plans or billing intervals?",
    answer:
      "Changing to another paid plan or switching between monthly and annual billing uses the current price for that selection. Prorated charges or credits may apply. Moving to the free plan takes effect at the end of your current paid billing period.",
  },
] as const;

export const EXISTING_SUBSCRIPTION_COPY =
  "Your existing per-user rate also applies to additional billable members. Plan limits and normal prorated charges still apply. Changing plans or billing intervals uses current pricing.";
