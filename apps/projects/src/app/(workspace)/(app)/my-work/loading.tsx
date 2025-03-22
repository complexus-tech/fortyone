import { Flex, Box, Text, Blur } from "ui";
import { ComplexusLogo } from "@/components/ui";

export default function Loading() {
  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Blur className="absolute left-1/2 right-1/2 z-[10] h-[400px] w-[400px] -translate-x-1/2 bg-warning/[0.07]" />
      <Flex align="center" direction="column" justify="center">
        <Box className="aspect-square w-max animate-pulse rounded-full bg-primary p-4">
          <ComplexusLogo className="h-8 text-white" />
        </Box>
        <Text className="mt-4" color="muted" fontWeight="medium">
          Loading workspace...
        </Text>
      </Flex>
    </Flex>
  );
}
