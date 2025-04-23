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
          teams, custom terminology and more.
        </Text>
        <Button color="tertiary">Manage subscription</Button>
      </Flex>
      <Divider className="mb-6 mt-4" />
      <Plans />
    </Box>
  );
};
