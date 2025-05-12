"use client";

import { Box, Button, Divider, Flex, Text } from "ui";
import { toast } from "sonner";
import { useState } from "react";
import { usePathname } from "next/navigation";
import { format } from "date-fns";
import { manageBilling } from "@/lib/actions/billing/manage-billing";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useMembers } from "@/lib/hooks/members";
import { useInvoices } from "@/lib/hooks/billing/invoices";
import { Plans } from "./components/plans";

// Define interface for invoice data based on the useInvoices hook return type
interface Invoice {
  id?: string;
  date?: string;
  amount?: string;
}

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
        <Text as="h1" className="mb-3 text-2xl font-medium">
          Billing
        </Text>
        <Flex className="mb-6" justify="between">
          <Text color="muted">
            For questions about billing,{" "}
            <Button className="p-0" size={null} variant="outline">
              contact us
            </Button>
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

  // Get plan display name
  const getPlanName = () => {
    if (tier === "business") return "Business";
    if (tier === "enterprise") return "Enterprise";
    return "Pro";
  };

  // Get plan price
  const getPlanPrice = () => {
    const isYearly = billingInterval === "year";
    if (tier === "business") {
      return isYearly ? "10" : "16";
    }
    return isYearly ? "7" : "12";
  };

  // Show paid tier users the subscription info and invoices
  return (
    <Box>
      <Text as="h1" className="mb-3 text-2xl font-medium">
        Billing
      </Text>
      <Flex className="mb-6" justify="between">
        <Text color="muted">
          For questions about billing,{" "}
          <Button className="p-0" size={null} variant="outline">
            contact us
          </Button>
        </Text>
        <Button
          color="tertiary"
          onClick={() => {
            setShowPlans(true);
          }}
          variant="outline"
        >
          All plans →
        </Button>
      </Flex>

      <Flex direction="column" gap={6}>
        {/* Current Subscription Info Card */}
        <Box className="rounded-lg border p-6">
          <Flex direction="column" gap={4}>
            <Flex align="start" justify="between">
              <Box>
                <Text as="h2" className="text-lg font-medium">
                  {getPlanName()}{" "}
                  {billingInterval === "year" ? "Yearly" : "Monthly"}
                  <Text
                    as="span"
                    className="text-muted-foreground bg-muted ml-2 rounded-full px-2 py-1 text-xs"
                  >
                    Current plan
                  </Text>
                </Text>
                <Text className="text-muted-foreground">
                  ${getPlanPrice()} per user/mo
                </Text>
              </Box>
              <Flex align="end" direction="column">
                <Flex align="center" gap={2}>
                  <Text className="text-muted-foreground">Users</Text>
                  <Text className="font-medium">{totalValidMembers}</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <Text className="text-muted-foreground">Next renewal</Text>
                  <Text className="font-medium">
                    <Text
                      as="span"
                      className="text-muted-foreground line-through"
                    >
                      $80
                    </Text>{" "}
                    $0 on May 25
                  </Text>
                </Flex>
              </Flex>
            </Flex>

            <Box className="bg-muted rounded-md p-4">
              <Flex align="center" gap={2}>
                <Box className="text-muted-foreground">
                  <svg
                    fill="none"
                    height="20"
                    viewBox="0 0 24 24"
                    width="20"
                    xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                      d="M12 8V12L15 15M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z"
                      stroke="currentColor"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth="2"
                    />
                  </svg>
                </Box>
                <Text className="text-muted-foreground">Free of charge</Text>
                <Text className="ml-auto">Valid through Oct 23</Text>
              </Flex>
            </Box>

            <Divider />

            <Flex className="flex-wrap gap-y-4">
              <Box className="w-1/2 pr-4">
                <Flex direction="column" gap={1}>
                  {renderFeatureItem("SAML authentication")}
                  {renderFeatureItem("Audit log")}
                  {renderFeatureItem("Third-party app management")}
                  {renderFeatureItem("Linear Asks")}
                </Flex>
              </Box>
              <Box className="w-1/2 pl-4">
                <Flex direction="column" gap={1}>
                  {renderFeatureItem("Domain claiming")}
                  {renderFeatureItem("Data warehouse sync")}
                  {renderFeatureItem("Issue SLAs")}
                  {renderFeatureItem("Priority support")}
                </Flex>
              </Box>
            </Flex>

            <Button
              className="mt-2 self-start"
              color="tertiary"
              disabled={isLoading}
              onClick={handleManageBilling}
            >
              Manage subscription
            </Button>
          </Flex>
        </Box>

        {/* Recent Invoices Card */}
        <Box className="rounded-lg border p-6">
          <Text as="h2" className="mb-6 text-lg font-medium">
            Recent invoices
          </Text>

          {invoices.length === 0 ? (
            <Text color="muted">No invoices available</Text>
          ) : (
            <Flex direction="column" gap={4}>
              {invoices.slice(0, 5).map((invoice, index) => {
                const invoiceItem = invoice as Invoice;
                return (
                  <Flex
                    align="center"
                    className="py-2"
                    justify="between"
                    key={invoiceItem.id || index}
                  >
                    <Text>{formatDate(invoiceItem.date)}</Text>
                    <Text>${invoiceItem.amount || "0.00"}</Text>
                  </Flex>
                );
              })}
            </Flex>
          )}
        </Box>
      </Flex>
    </Box>
  );
};
