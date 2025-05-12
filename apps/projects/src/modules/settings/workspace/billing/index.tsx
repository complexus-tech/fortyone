"use client";

import { Badge, Box, Button, Divider, Flex, Text } from "ui";
import { toast } from "sonner";
import { useState } from "react";
import { usePathname } from "next/navigation";
import { format } from "date-fns";
import { ArrowRight2Icon, NewTabIcon } from "icons";
import { cn } from "lib";
import { manageBilling } from "@/lib/actions/billing/manage-billing";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useMembers } from "@/lib/hooks/members";
import { useInvoices } from "@/lib/hooks/billing/invoices";
import { RowWrapper } from "@/components/ui";
import { SectionHeader } from "../../components";
import { Plans } from "./components/plans";
import { plans, getPlanFeaturesList } from "./components/plan-data";

export const Billing = () => {
  const pathname = usePathname();
  const { data: members } = useMembers();
  const { data: invoices = [] } = useInvoices();
  const { tier, billingInterval } = useSubscriptionFeatures();
  const [isLoading, setIsLoading] = useState(false);
  const [showPlans, setShowPlans] = useState(tier === "free");
  const totalValidMembers =
    members?.filter((member) => member.role !== "guest").length || 0;

  const handleManageBilling = async () => {
    setIsLoading(true);
    const host = `${window.location.protocol}//${window.location.host}`;
    const returnUrl = `${host}${pathname}`;
    const res = await manageBilling(returnUrl);
    if (res.error?.message) {
      toast.error("Failed to manage billing", {
        description: res.error.message,
      });
      setIsLoading(false);
    }
    window.location.href = res.data!.url;
  };

  const formatDate = (date: string | Date | undefined) => {
    if (!date) return "";
    return format(new Date(date), "MMM dd, yyyy");
  };

  const renderFeatureItem = (text: string) => {
    return (
      <Flex align="center" gap={2}>
        <svg
          fill="none"
          height="16"
          viewBox="0 0 24 24"
          width="16"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M5 13L9 17L19 7"
            stroke="currentColor"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
          />
        </svg>
        <Text>{text}</Text>
      </Flex>
    );
  };

  // Show free tier users only the plans component
  if (tier === "free" || showPlans) {
    return (
      <Box>
        <Text as="h1" className="mb-4 text-2xl font-medium">
          Billing
        </Text>
        <Flex align="center" className="mb-6" justify="between">
          <Text color="muted">
            For questions about billing,{" "}
            <a
              className="text-dark dark:text-white"
              href="mailto:info@complexus.app"
            >
              contact us.
            </a>{" "}
            Your workspace has {totalValidMembers} users.
          </Text>
          {tier !== "free" && (
            <Button
              color="tertiary"
              onClick={() => {
                setShowPlans(false);
              }}
              variant="outline"
            >
              Back to subscription
            </Button>
          )}
        </Flex>
        <Plans />
      </Box>
    );
  }

  // Map tier to the name displayed in plans
  const getTierName = () => {
    const tierNameMap: Record<string, string> = {
      pro: "Pro",
      business: "Business",
      enterprise: "Enterprise",
      free: "Hobby",
      trial: "Trial",
    };
    return tierNameMap[tier] || "Pro";
  };

  // Find current plan based on tier name
  const currentPlan = plans.find((plan) => plan.name === getTierName());

  // Get pricing
  const getPlanPrice = () => {
    const isYearly = billingInterval === "year";

    if (tier === "business") {
      return isYearly ? "8" : "10";
    }
    if (tier === "pro") {
      return isYearly ? "5.6" : "7";
    }
    return "0";
  };

  // Get features from the current plan
  const planFeatures = currentPlan ? getPlanFeaturesList(currentPlan) : [];
  const halfFeatureLength = Math.ceil(planFeatures.length / 2);
  const firstColumnFeatures = planFeatures.slice(0, halfFeatureLength);
  const secondColumnFeatures = planFeatures.slice(halfFeatureLength);

  return (
    <Box
      className={cn("mx-auto max-w-[54rem]", {
        "max-w-7xl": showPlans,
      })}
    >
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Billing
      </Text>
      <Flex align="center" className="mb-4" justify="between">
        <Text color="muted">
          For questions about billing,{" "}
          <a
            className="text-dark dark:text-white"
            href="mailto:info@complexus.app"
          >
            contact us.
          </a>{" "}
          Your plan has {totalValidMembers} users.
        </Text>
        <Button
          color="tertiary"
          onClick={() => {
            setShowPlans(true);
          }}
          rightIcon={<ArrowRight2Icon />}
          variant="naked"
        >
          <span className="opacity-80">All plans</span>
        </Button>
      </Flex>

      <Box className="mb-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              color="tertiary"
              disabled={isLoading}
              onClick={handleManageBilling}
              variant="outline"
            >
              Manage subscription
            </Button>
          }
          description="Details about your current plan and subscription."
          title="Current Subscription"
        />
        <Box className="p-6">
          <Flex direction="column" gap={4}>
            <Flex align="start" justify="between">
              <Box>
                <Text
                  as="h3"
                  className="mb-0.5 flex items-center gap-2 font-medium"
                >
                  {getTierName()}{" "}
                  {billingInterval === "year" ? "Yearly" : "Monthly"}
                  <Badge
                    className="h-7 px-1.5 text-base"
                    color="tertiary"
                    size="lg"
                    variant="outline"
                  >
                    Current plan
                  </Badge>
                </Text>
                <Text color="muted">${getPlanPrice()} per user/mo</Text>
              </Box>
              <Box>
                <Text className="mb-0.5">Next renewal</Text>
                <Text color="muted">May 25</Text>
              </Box>
            </Flex>
            <Divider />
            <Flex className="flex-wrap gap-y-4">
              <Box className="w-1/2 pr-4">
                <Flex direction="column" gap={2}>
                  {firstColumnFeatures.map((feature) =>
                    renderFeatureItem(feature),
                  )}
                </Flex>
              </Box>
              <Box className="w-1/2 pl-4">
                <Flex direction="column" gap={2}>
                  {secondColumnFeatures.map((feature) =>
                    renderFeatureItem(feature),
                  )}
                </Flex>
              </Box>
            </Flex>
          </Flex>
        </Box>
      </Box>

      {/* Recent Invoices Card */}
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Your recent billing history and invoices."
          title="Recent Invoices"
        />
        {invoices.length !== 0 ? (
          <Flex className="px-4 py-5 md:px-6">
            <Text color="muted">No invoices available</Text>
          </Flex>
        ) : (
          <>
            {invoices.map(
              ({ stripeInvoiceId, hostedUrl, invoiceDate, amountPaid }) => (
                <RowWrapper
                  className="px-4 last-of-type:border-b-0 md:px-6"
                  key={stripeInvoiceId}
                >
                  <Text>{formatDate(invoiceDate)}</Text>
                  <Text className="font-semibold" color="muted">
                    {new Intl.NumberFormat("en-US", {
                      style: "currency",
                      currency: "USD",
                    }).format(amountPaid || 0)}
                  </Text>
                  {hostedUrl ? (
                    <Button
                      color="tertiary"
                      href={hostedUrl}
                      rightIcon={<NewTabIcon className="h-[1.125rem]" />}
                      target="_blank"
                      variant="naked"
                    >
                      View
                    </Button>
                  ) : null}
                </RowWrapper>
              ),
            )}
          </>
        )}
      </Box>
    </Box>
  );
};
