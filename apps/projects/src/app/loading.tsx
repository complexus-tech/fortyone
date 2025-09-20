import { Flex, Text } from "ui";
import { FortyOneLogo } from "@/components/ui";

export default function Loading() {
  return (
    <Flex
      align="center"
      className="relative h-dvh dark:bg-black"
      justify="center"
    >
      <Flex align="center" direction="column" justify="center">
        <FortyOneLogo className="animate-pulse" />
        <Text className="mt-4" color="muted" fontWeight="medium">
          Loading workspace...
        </Text>
      </Flex>
    </Flex>
  );
}
