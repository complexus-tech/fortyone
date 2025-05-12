"use client";
import { Flex, Text, Box, Button } from "ui";
import { ErrorIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { motion } from "framer-motion";
import { useState } from "react";
import { toast } from "sonner";
import { usePathname, useRouter } from "next/navigation";
import type { Plan as CheckoutPlan } from "@/lib/actions/billing/checkout"; // Renaming to avoid conflict
import { checkout } from "@/lib/actions/billing/checkout";
import {
  changePlan,
  type Plan as ChangePlanPlan,
} from "@/lib/actions/billing/change-plan"; // Correct import and type
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { plans, featureLabels } from "./plan-data";

const FeatureCheck = ({ available }: { available: boolean }) => (
  <Box className="flex">
    {available ? (
      <SuccessIcon className="text-primary dark:text-primary" />
    ) : (
      <ErrorIcon className="dark:text-white" />
    )}
  </Box>
);

// Helper to render a feature value
const FeatureValue = ({
  value,
}: {
  value: boolean | string | undefined | null;
}) => {
  if (value === undefined || value === null) {
    return <ErrorIcon className="dark:text-white" />;
  }

  if (typeof value === "boolean") {
    return <FeatureCheck available={value} />;
  }

  return <Text>{value}</Text>;
};

type Billing = "annual" | "monthly";

// Combined Plan type for handlePlanAction
type ActionPlan = CheckoutPlan | ChangePlanPlan;

export const Plans = () => {
  const { tier, billingInterval } = useSubscriptionFeatures();
  const router = useRouter();
  const pathname = usePathname(); // Used for checkout's cancelUrl
  const [isProLoading, setIsProLoading] = useState(false);
  const [isBusinessLoading, setIsBusinessLoading] = useState(false);
  const [billing, setBilling] = useState<Billing>(
    billingInterval === "year" ? "annual" : "monthly",
  );
  const proPrice = 7;
  const businessPrice = 10;

  const handlePlanAction = async (plan: ActionPlan) => {
    const host = `${window.location.protocol}//${window.location.host}`;
    const cancelUrl = `${host}${pathname}`; // For checkout
    const successUrl = `${host}/my-work`; // For checkout

    if (plan.startsWith("pro")) {
      setIsProLoading(true);
    } else if (plan.startsWith("business")) {
      setIsBusinessLoading(true);
    }

    let response;
    try {
      if (tier === "free" || tier === "trial") {
        // Checkout requires successUrl and cancelUrl
        response = await checkout({
          plan: plan as CheckoutPlan,
          successUrl,
          cancelUrl,
        });
        if (response.data?.url) {
          window.location.href = response.data.url;
        }
      } else {
        // changePlan only takes the plan type
        response = await changePlan(plan as ChangePlanPlan);
        // After changePlan, we might need to refresh or show a success message
        // Assuming changePlan's response will guide this or it's handled by a page refresh/global state update
        if (!response.error) {
          toast.success("Plan change initiated successfully!");
          // Potentially refresh data or rely on Stripe webhook to update UI
          // For now, a toast message will confirm initiation.
          router.refresh(); // Or navigate to a confirmation page, or update state
        }
      }

      if (response.error?.message) {
        toast.error("Failed to process plan change", {
          description: response.error.message,
        });
      }
      // Success for checkout is handled by redirect.
      // Success for changePlan is handled above with a toast and potential reload.
    } catch (error) {
      toast.error("An unexpected error occurred.", {
        description: error instanceof Error ? error.message : String(error),
      });
    } finally {
      setIsProLoading(false);
      setIsBusinessLoading(false);
    }
  };

  // Determine Pro Button Text and Disabled State
  let proButtonText = "Upgrade";
  let isProButtonDisabled = isProLoading;

  if (tier === "pro") {
    const currentProIntervalIsYear = billingInterval === "year";
    const targetIntervalIsAnnual = billing === "annual";

    if (targetIntervalIsAnnual && !currentProIntervalIsYear) {
      proButtonText = "Switch to yearly";
      isProButtonDisabled = isProLoading;
    } else if (!targetIntervalIsAnnual && currentProIntervalIsYear) {
      proButtonText = "Switch to monthly";
      isProButtonDisabled = isProLoading;
    } else {
      proButtonText = "Current plan";
      isProButtonDisabled = true;
    }
  } else if (tier === "business" || tier === "enterprise") {
    proButtonText = "Downgrade"; // To Pro
    isProButtonDisabled = isProLoading;
  }
  if (isProLoading) isProButtonDisabled = true;

  // Determine Business Button Text and Disabled State
  let businessButtonText = "Upgrade";
  let isBusinessButtonDisabled = isBusinessLoading;

  if (tier === "business") {
    const currentBusinessIntervalIsYear = billingInterval === "year";
    const targetIntervalIsAnnual = billing === "annual";

    if (targetIntervalIsAnnual && !currentBusinessIntervalIsYear) {
      businessButtonText = "Switch to yearly";
      isBusinessButtonDisabled = isBusinessLoading;
    } else if (!targetIntervalIsAnnual && currentBusinessIntervalIsYear) {
      businessButtonText = "Switch to monthly";
      isBusinessButtonDisabled = isBusinessLoading;
    } else {
      businessButtonText = "Current plan";
      isBusinessButtonDisabled = true;
    }
  } else if (tier === "enterprise") {
    businessButtonText = "Downgrade"; // To Business
    isBusinessButtonDisabled = isBusinessLoading;
  }
  if (isBusinessLoading) isBusinessButtonDisabled = true;

  return (
    <Box className="hidden overflow-x-auto md:block">
      <Box>
        <Flex className="border-b border-gray-100 dark:border-dark-100">
          <Box className="flex w-1/5 items-end px-3 py-6">
            <motion.div
              initial={{ y: 20, opacity: 0 }}
              transition={{
                duration: 1,
                delay: 0.3,
              }}
              viewport={{ once: true, amount: 0.5 }}
              whileInView={{ y: 0, opacity: 1 }}
            >
              <Box className="flex w-full gap-1 rounded-[0.7rem] border border-gray-100 bg-gray-50 p-1 dark:border-dark-100 dark:bg-dark-300">
                {["annual", "monthly"].map((option) => (
                  <Button
                    className={cn("px-4 capitalize", {
                      "opacity-80": option !== billing,
                    })}
                    color={option === billing ? "primary" : "tertiary"}
                    key={option}
                    onClick={() => {
                      setBilling(option as Billing);
                    }}
                    size="sm"
                    variant={option === billing ? "solid" : "naked"}
                  >
                    {option}
                  </Button>
                ))}
              </Box>
            </motion.div>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Hobby</Text>
            <Text className="mb-2 text-3xl font-semibold">$0</Text>
            <Button
              align="center"
              color="tertiary"
              disabled={tier === "free" || tier === "trial"}
              fullWidth
            >
              {tier === "free" || tier === "trial"
                ? "Continue free"
                : "Downgrade"}
            </Button>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Pro</Text>
            <Text className="mb-2 text-3xl font-semibold">
              ${billing === "annual" ? (proPrice * 0.8).toFixed(2) : proPrice}
              <Text as="span" color="muted" fontSize="md" fontWeight="medium">
                {" "}
                per user/month
              </Text>
            </Text>
            <Button
              align="center"
              color="tertiary"
              disabled={isProButtonDisabled}
              fullWidth
              onClick={() => {
                const planToSubmit: ActionPlan =
                  billing === "annual" ? "pro_yearly" : "pro_monthly";
                handlePlanAction(planToSubmit);
              }}
            >
              {proButtonText}
            </Button>
          </Box>
          <Box className="w-1/5 rounded-t-2xl border border-b-0 border-gray-100 bg-gray-50 px-4 py-6 dark:border-dark-100 dark:bg-dark-300">
            <Text className="mb-2 text-2xl">Business</Text>
            <Text className="mb-2 text-3xl font-semibold">
              ${billing === "annual" ? businessPrice * 0.8 : businessPrice}
              <Text as="span" color="muted" fontSize="md" fontWeight="medium">
                {" "}
                per user/month
              </Text>
            </Text>
            <Button
              align="center"
              color="tertiary"
              disabled={isBusinessButtonDisabled}
              fullWidth
              onClick={() => {
                const planToSubmit: ActionPlan =
                  billing === "annual" ? "business_yearly" : "business_monthly";
                handlePlanAction(planToSubmit);
              }}
            >
              {businessButtonText}
            </Button>
          </Box>
          <Box className="w-1/5 px-4 py-6">
            <Text className="mb-2 text-2xl">Enterprise</Text>
            <Text className="mb-2 text-3xl font-semibold">Custom</Text>
            <Button
              align="center"
              color="tertiary"
              disabled={tier === "enterprise"}
              fullWidth
              href="mailto:info@complexus.app"
            >
              {tier === "enterprise" ? "Current plan" : "Contact sales"}
            </Button>
          </Box>
        </Flex>

        {/* Limits section */}
        <Box>
          <Flex className="border-b border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300/50">
            <Box className="w-1/5 px-4 py-4">
              <Text fontWeight="semibold">Limits</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.name === "Business",
                })}
                key={plan.name}
              />
            ))}
          </Flex>
          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">Members</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn(
                  "w-1/5 px-4 py-4",
                  plan.highlighted &&
                    "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300",
                )}
                key={`${plan.name}-members`}
              >
                <Text>{plan.limits.members}</Text>
              </Box>
            ))}
          </Flex>

          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">File uploads</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.highlighted,
                })}
                key={`${plan.name}-files`}
              >
                <Text>{plan.limits.fileUploads}</Text>
              </Box>
            ))}
          </Flex>

          <Flex className="border-b border-gray-100 dark:border-dark-100">
            <Box className="w-1/5 px-4 py-4">
              <Text className="opacity-80">Stories</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.highlighted,
                })}
                key={`${plan.name}stories`}
              >
                <Text>{plan.limits.issues}</Text>
              </Box>
            ))}
          </Flex>
        </Box>

        {/* Features section */}
        <Box>
          <Flex className="border-b border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300/50">
            <Box className="w-1/5 px-4 py-4">
              <Text fontWeight="semibold">Features</Text>
            </Box>
            {plans.map((plan) => (
              <Box
                className={cn("w-1/5 px-4 py-4", {
                  "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                    plan.name === "Business",
                })}
                key={plan.name}
              />
            ))}
          </Flex>

          {/* Generate all possible feature rows */}
          {Object.entries(featureLabels).map(
            ([featureKey, featureLabel], index, array) => {
              // Skip 'teams' feature since it's already in limits section
              if (featureKey === "teams") return null;

              const isLastItem = index === array.length - 1;

              return (
                <Flex
                  className={
                    !isLastItem
                      ? "border-b border-gray-100 dark:border-dark-100"
                      : ""
                  }
                  key={featureKey}
                >
                  <Box className="w-1/5 px-4 py-4">
                    <Text className="opacity-80">{featureLabel}</Text>
                  </Box>
                  {plans.map((plan) => {
                    const value =
                      plan.features[featureKey as keyof typeof plan.features];
                    const isHighlighted = plan.highlighted;
                    return (
                      <Box
                        className={cn("w-1/5 px-4 py-4", {
                          "border-x border-gray-100 bg-gray-50 dark:border-dark-100 dark:bg-dark-300":
                            isHighlighted,
                        })}
                        key={`${plan.name}-${featureKey}`}
                      >
                        <FeatureValue value={value} />
                      </Box>
                    );
                  })}
                </Flex>
              );
            },
          )}
        </Box>
      </Box>
    </Box>
  );
};
