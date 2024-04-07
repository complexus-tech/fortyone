import { Button, Flex, Text } from "ui";
import { ArrowRightIcon } from "icons";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="pt-20">
      <Flex
        align="center"
        className="mb-8 mt-20 text-center"
        direction="column"
      >
        <Button
          className="px-3 text-sm md:text-base"
          color="tertiary"
          rightIcon={
            <ArrowRightIcon className="relative top-[0.5px] h-3 w-auto" />
          }
          rounded="full"
          size="sm"
        >
          Announcing Early Adopters Plan
        </Button>

        <Text
          as="h1"
          className="mt-6 h-max max-w-6xl pb-2 text-4xl md:text-7xl"
          color="gradient"
          fontWeight="medium"
        >
          Empowering teams to conquer project complexity.
        </Text>
        <Text className="mt-8 max-w-[600px] md:text-lg" color="muted">
          Revolutionize project management. Simplify workflows, enhance
          collaboration, achieve exceptional results.
        </Text>
        <Flex align="center" className="mt-10" gap={4}>
          <Button
            className="border border-primary"
            rounded="full"
            size="lg"
            variant="outline"
          >
            Talk to us
          </Button>
          <Button rounded="full" size="lg">
            Get Early Access
          </Button>
        </Flex>
        <Text className="mt-8" color="muted" fontSize="sm">
          No credit card required.
        </Text>
      </Flex>
    </Container>
  );
};
