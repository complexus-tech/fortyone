import { Flex, Box, Text } from "ui";
import { FortyOneLogo } from "@/components/ui";

export default function Loading() {
  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Box className="aspect-square w-max animate-pulse rounded-full bg-primary p-4">
          <FortyOneLogo className="h-8 text-white" />
        </Box>
        <Text className="mt-4" color="muted" fontWeight="medium">
          Loading workspace...
        </Text>
      </Flex>
    </Flex>
  );
}
