"use client";

import { Box, Button, Divider, Flex, Text } from "ui";
import { toast } from "sonner";
import { useState } from "react";
import { usePathname } from "next/navigation";
import { manageBilling } from "@/lib/actions/billing/manage-billing";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useMembers } from "@/lib/hooks/members";
import { useInvoices } from "@/lib/hooks/billing/invoices";
import { Plans } from "./components/plans";

export const Billing = () => {
  const pathname = usePathname();
  const { data: members } = useMembers();
  const { data: invoices = [] } = useInvoices();
  const { tier } = useSubscriptionFeatures();
  const [isLoading, setIsLoading] = useState(false);
  const totalValidMembers = members?.filter(
    (member) => member.role !== "guest",
  ).length;

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

  return (
    <Box>
      <Text as="h1" className="mb-3 text-2xl font-medium">
        Billing and plans
      </Text>
      <Flex justify="between">
        {(tier === "free" || tier === "trial") && (
          <Text className="max-w-xl" color="muted">
            Scale your workspace as your team grows. Currently on{" "}
            <Text as="span" fontWeight="semibold" transform="capitalize">
              {tier}
            </Text>{" "}
            plan with {totalValidMembers} members - upgrade to unlock
            Objectives, OKRs, Private teams, and more.
          </Text>
        )}
        {tier === "pro" && (
          <Text className="max-w-xl" color="muted">
            You&apos;re on the{" "}
            <Text as="span" fontWeight="semibold" transform="capitalize">
              {tier}
            </Text>{" "}
            plan with {totalValidMembers} members. Need more? Upgrade to
            Business for unlimited teams, private teams, custom terminology and
            more.
          </Text>
        )}
        {(tier === "business" || tier === "enterprise") && (
          <Text className="max-w-xl" color="muted">
            You&apos;re on the{" "}
            <Text as="span" fontWeight="semibold" transform="capitalize">
              {tier}
            </Text>{" "}
            plan. {totalValidMembers} members, unlimited objectives, private
            teams, custom terminology, and more.
          </Text>
        )}

        <Button
          color="tertiary"
          disabled={isLoading || ["trial", "free"].includes(tier)}
          onClick={handleManageBilling}
        >
          Manage subscription
        </Button>
      </Flex>
      <Divider className="mb-6 mt-4" />
      <Plans />
    </Box>
  );
};
