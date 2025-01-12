"use client";

import { Box, Flex, Text, Button, Badge } from "ui";
import { SectionHeader } from "../components";

export const BillingSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Billing Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<Button color="primary">Upgrade Plan</Button>}
          description="Manage your subscription and billing details."
          title="Current Plan"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Flex align="center" justify="between">
              <Box>
                <Flex align="center" gap={2}>
                  <Text className="font-medium">Free Plan</Text>
                  <Badge color="tertiary">Current Plan</Badge>
                </Flex>
                <Text className="mt-1" color="muted">
                  Basic features for small teams
                </Text>
              </Box>
              <Text className="text-xl font-semibold">$0/month</Text>
            </Flex>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Manage your payment methods and billing information."
          title="Payment Method"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No payment method added yet.</Text>
              <Button className="mt-4" color="tertiary" variant="outline">
                Add Payment Method
              </Button>
            </Box>
          </Flex>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="View your past invoices and billing history."
          title="Billing History"
        />

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Text color="muted">No billing history available.</Text>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
