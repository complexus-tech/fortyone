import { Button, Flex, Text } from "ui";
import { Blur, Container } from "@/components/ui";

export const CallToAction = () => {
  return (
    <Container className="relative max-w-7xl py-28">
      <Flex
        align="center"
        className="md:mt-18 mb-8 mt-16 text-center"
        direction="column"
      >
        <Text
          as="h1"
          className="mt-6 h-max max-w-6xl pb-2 text-7xl"
          color="gradient"
          fontWeight="medium"
        >
          Empowering teams to conquer project complexity.
        </Text>
        <Text
          className="mt-4 max-w-[600px] md:mt-6"
          color="muted"
          fontSize="lg"
        >
          Revolutionize project management. Simplify workflows, enhance
          collaboration, achieve exceptional results.
        </Text>
        <Flex align="center" className="mt-8" gap={3}>
          <Button rounded="full" size="lg">
            Get started for free
          </Button>
          <Button
            className="border border-primary"
            rounded="full"
            size="lg"
            variant="outline"
          >
            Sign in
          </Button>
        </Flex>
      </Flex>
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[400px] w-[400px] -translate-x-1/2 -translate-y-1/2 bg-primary/40 dark:bg-warning/5" />
    </Container>
  );
};
