"use client";
import { Flex, Text, Box, Button } from "ui";
import { ErrorIcon, SuccessIcon } from "icons";
import { cn } from "lib";
import { motion } from "framer-motion";
import { useState } from "react";
import { toast } from "sonner";
import { usePathname } from "next/navigation";
import { checkout } from "@/lib/actions/billing/checkout";
import { changePlan } from "@/lib/actions/billing/change-plan";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import type { Plan } from "@/lib/actions/billing/types";
import { ConfirmDialog } from "@/components/ui";
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

type Tier = ReturnType<typeof useSubscriptionFeatures>["tier"];
type BillingIntervalApi = ReturnType<
  typeof useSubscriptionFeatures
>["billingInterval"];

const getProButtonState = (
  currentTier: Tier,
  currentIntervalApi: BillingIntervalApi,
  selectedIntervalUi: Billing,
  isLoading: boolean,
): { text: string; disabled: boolean } => {
  let text = "Upgrade";
  let disabled = isLoading;

  if (currentTier === "pro") {
    const currentProIntervalIsYear = currentIntervalApi === "year";
    const targetIntervalIsAnnual = selectedIntervalUi === "annual";

    if (targetIntervalIsAnnual && !currentProIntervalIsYear) {
      text = "Switch to yearly";
      disabled = isLoading;
    } else if (!targetIntervalIsAnnual && currentProIntervalIsYear) {
      text = "Switch to monthly";
      disabled = isLoading;
    } else {
      text = "Current plan";
      disabled = true;
    }
  } else if (currentTier === "business" || currentTier === "enterprise") {
    text = "Downgrade";
    disabled = isLoading;
  }

  if (isLoading) {
    disabled = true;
  }
  return { text, disabled };
};

const getBusinessButtonState = (
  currentTier: Tier,
  currentIntervalApi: BillingIntervalApi,
  selectedIntervalUi: Billing,
  isLoading: boolean,
): { text: string; disabled: boolean } => {
  let text = "Upgrade";
  let disabled = isLoading;

  if (currentTier === "business") {
    const currentBusinessIntervalIsYear = currentIntervalApi === "year";
    const targetIntervalIsAnnual = selectedIntervalUi === "annual";

    if (targetIntervalIsAnnual && !currentBusinessIntervalIsYear) {
      text = "Switch to yearly";
      disabled = isLoading;
    } else if (!targetIntervalIsAnnual && currentBusinessIntervalIsYear) {
      text = "Switch to monthly";
      disabled = isLoading;
    } else {
      text = "Current plan";
      disabled = true;
    }
  } else if (currentTier === "enterprise") {
    text = "Downgrade";
    disabled = isLoading;
  }

  if (isLoading) {
    disabled = true;
  }
  return { text, disabled };
};

export const Plans = () => {
  const { tier, billingInterval } = useSubscriptionFeatures();
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);
  const [isProLoading, setIsProLoading] = useState(false);
  const [isBusinessLoading, setIsBusinessLoading] = useState(false);
  const [billing, setBilling] = useState<Billing>(
    billingInterval === "year" ? "annual" : "monthly",
  );
  const [pendingAction, setPendingAction] = useState<{
    plan: Plan;
    type: "upgrade" | "downgrade" | "switch";
    from: string;
    to: string;
  } | null>(null);
  const proPrice = 7;
  const businessPrice = 10;

  const handlePlanAction = async (plan: Plan) => {
    const host = `${window.location.protocol}//${window.location.host}`;
    const cancelUrl = `${host}${pathname}`;
    const successUrl = `${host}/my-work`;

    if (plan.startsWith("pro")) {
      setIsProLoading(true);
    } else if (plan.startsWith("business")) {
      setIsBusinessLoading(true);
    }

    try {
      if (tier === "free" || tier === "trial") {
        const response = await checkout({
          plan,
          successUrl,
          cancelUrl,
        });
        if (response.error?.message) {
          toast.error("Failed to upgrade plan", {
            description: response.error.message,
          });
          return;
        }
        window.location.href = response.data?.url ?? "";
      } else {
        const toastId = toast.loading("Processing plan change...", {
          description: "Please wait while we process your request.",
        });
        const response = await changePlan(plan);
        if (response.error?.message) {
          toast.error("Failed to process plan change", {
            description: response.error.message,
          });
          return;
        }
        toast.success("Plan change initiated!", {
          description:
            "You will receive an email when the plan change is complete.",
          id: toastId,
        });
      }
    } catch (error) {
      toast.error("An unexpected error occurred.", {
        description: error instanceof Error ? error.message : String(error),
      });
    } finally {
      setIsProLoading(false);
      setIsBusinessLoading(false);
    }
  };

  const getTierLabel = (tierName: Tier): string => {
    const labels: Record<Tier, string> = {
      free: "Hobby",
      trial: "Trial",
      pro: "Pro",
      business: "Business",
      enterprise: "Enterprise",
    };
    return labels[tierName];
  };

  const getActionType = (
    currentTier: Tier,
    targetTier: string,
  ): "upgrade" | "downgrade" | "switch" => {
    if (currentTier === "free" || currentTier === "trial") return "upgrade";

    const tierOrder = { free: 0, trial: 0, pro: 1, business: 2, enterprise: 3 };
    const currentTierValue = tierOrder[currentTier];
    const targetTierValue = tierOrder[targetTier as Tier] || 0;

    if (currentTier === targetTier.toLowerCase()) {
      return "switch";
    }

    return currentTierValue > targetTierValue ? "downgrade" : "upgrade";
  };

  const getDialogProps = () => {
    if (!pendingAction) return { title: "", description: "" };

    switch (pendingAction.type) {
      case "upgrade":
        return {
          title: `Upgrade to ${pendingAction.to}`,
          description: `Your current ${pendingAction.from} plan will be upgraded to ${pendingAction.to}, and we will charge you the price difference to your current payment method.`,
        };
      case "downgrade":
        return {
          title: `Downgrade to ${pendingAction.to}`,
          description: `Your current ${pendingAction.from} plan will be downgraded to ${pendingAction.to}. You will lose access to all ${pendingAction.from} features.`,
        };
      case "switch":
        return {
          title: `Switch to ${pendingAction.to}`,
          description: `Your current ${pendingAction.from} plan will be switched to ${pendingAction.to}. Your billing cycle will be updated accordingly.`,
        };
    }
  };

  const { text: proButtonText, disabled: isProButtonDisabled } =
    getProButtonState(tier, billingInterval, billing, isProLoading);

  const { text: businessButtonText, disabled: isBusinessButtonDisabled } =
    getBusinessButtonState(tier, billingInterval, billing, isBusinessLoading);

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
              onClick={() => {
                if (tier !== "free" && tier !== "trial") {
                  setPendingAction({
                    plan: "free" as Plan,
                    type: "downgrade",
                    from: getTierLabel(tier),
                    to: "Hobby",
                  });
                  setIsOpen(true);
                }
              }}
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
                const planId =
                  billing === "annual" ? "pro_yearly" : "pro_monthly";
                const actionType = getActionType(tier, "pro");

                setPendingAction({
                  plan: planId,
                  type: actionType,
                  from: getTierLabel(tier),
                  to: `Pro ${billing === "annual" ? "yearly" : "monthly"}`,
                });
                setIsOpen(true);
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
                const planId =
                  billing === "annual" ? "business_yearly" : "business_monthly";
                const actionType = getActionType(tier, "business");

                setPendingAction({
                  plan: planId,
                  type: actionType,
                  from: getTierLabel(tier),
                  to: `Business ${billing === "annual" ? "yearly" : "monthly"}`,
                });
                setIsOpen(true);
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
      <ConfirmDialog
        {...getDialogProps()}
        isOpen={isOpen}
        onClose={() => {
          setIsOpen(false);
          setPendingAction(null);
        }}
        onConfirm={() => {
          if (pendingAction) {
            handlePlanAction(pendingAction.plan);
          }
          setIsOpen(false);
          setPendingAction(null);
        }}
      />
    </Box>
  );
};
