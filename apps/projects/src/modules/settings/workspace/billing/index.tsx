import { Box, Button, Divider, Flex, Text } from "ui";
import { Plans } from "./components/plans";

export const Billing = () => {
  return (
    <Box>
      <Text as="h1" className="mb-3 text-2xl font-medium">
        Billing and plans
      </Text>
      <Flex justify="between">
        {/* <Text className="max-w-xl" color="muted">
          Scale your workspace as your team grows. Currently on{" "}
          <Text as="span" fontWeight="semibold">
            Hobby
          </Text>{" "}
          - upgrade to unlock Objectives, OKRs, and more.
        </Text> */}

        <Text className="max-w-xl" color="muted">
          You&apos;re on the{" "}
          <Text as="span" fontWeight="semibold">
            Pro
          </Text>{" "}
          plan. Need more? Upgrade to Business for unlimited teams, private
          teams, and custom terminology.
        </Text>
        <Button color="tertiary">Manage subscription</Button>
      </Flex>
      <Divider className="mb-6 mt-4" />
      <Plans />
    </Box>
  );
};

// Alternative copy options:

// Option 1 - More conversational:
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Subscription & Billing
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   You're currently on our Hobby plan. Ready to unlock more power? Upgrade to a paid plan for advanced features.
// </Text>

// Option 2 - Value-focused:
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Plans & Pricing
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   Scale your workspace as your team grows. Currently on Hobby tier (free) - upgrade to unlock Objectives, OKRs, and more.
// </Text>

// Option 3 - Feature-focused:
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Upgrade Your Workspace
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   Your free Hobby plan includes basic features. Unlock custom workflows, unlimited storage, and priority support with our paid plans.
// </Text>

// Option 4 - Growth-oriented:
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Workspace Plans
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   As your team evolves, so should your tools. Upgrade from your current Hobby plan to unlock premium features designed for growing teams.
// </Text>

// Option 5 - Direct and simple:
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Select a Plan
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   Currently using Hobby (free). Choose a plan that best fits your team's needs with transparent pricing.
// </Text>

// For Pro plan users:

// Option 2 - Value-focused (Pro):
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Plans & Pricing
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   You're on the Pro plan. Need more? Upgrade to Business for unlimited teams, private teams, and custom terminology.
// </Text>

// Option 3 - Feature-focused (Pro):
// <Text as="h1" className="mb-3 text-2xl font-medium">
//   Upgrade Your Workspace
// </Text>
// <Text className="mb-6 max-w-xl" color="muted">
//   Your Pro plan includes up to 3 teams and basic OKR tracking. Scale to Business for unlimited objectives, private teams, and priority support.
// </Text>
