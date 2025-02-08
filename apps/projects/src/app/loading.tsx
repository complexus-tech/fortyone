import { Flex, Box, Text } from "ui";
import { ComplexusLogo } from "@/components/ui";

export default function Loading(): JSX.Element {
  return (
    <Flex align="center" className="h-screen dark:bg-black" justify="center">
      <Flex align="center" direction="column" justify="center">
        <Box className="aspect-square w-max animate-pulse rounded-full bg-primary p-4">
          <ComplexusLogo className="h-8 text-white" />
        </Box>
        <Text className="mt-4" fontSize="lg" fontWeight="medium">
          Loading workspace...
        </Text>
      </Flex>
    </Flex>
  );
}
