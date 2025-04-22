import { Box, Text } from "ui";
import { Plans } from "./components/plans";

export const Billing = () => {
  return (
    <Box>
      <Text as="h1" className="mb-3 text-2xl font-medium">
        Billing and plans
      </Text>
      <Text className="mb-6 max-w-xl" color="muted">
        Your Workspace is currently on the Hobby Plan. Upgrade to a paid plan to
        unlock all features.
      </Text>
      <Plans />
    </Box>
  );
};
