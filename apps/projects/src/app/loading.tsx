import { Flex, Text } from "ui";
import { Logo } from "@/components/ui";

export default function Loading() {
  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <Logo className="animate-pulse" asIcon />
        <Text className="mt-4" color="muted" fontWeight="medium">
          Loading workspace...
        </Text>
      </Flex>
    </Flex>
  );
}
