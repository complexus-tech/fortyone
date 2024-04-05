import { Badge, Box, Button, Flex, Text } from "ui";
import { ArrowRightIcon } from "icons";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/5" />
      <Container className="relative pt-20">
        <Flex
          align="center"
          className="md:mt-18 mb-8 mt-16 text-center"
          direction="column"
        >
          <Badge
            className="border-primary/30 bg-primary/10 text-dark-100 dark:border-primary/15 dark:bg-primary/10 dark:text-gray-200"
            rounded="full"
            size="lg"
          >
            Announcing Early Adopters Plan
            <ArrowRightIcon className="h-3 w-auto" />
          </Badge>

          <Text
            as="h1"
            className="mt-6 h-max max-w-6xl pb-2 text-7xl"
            color="gradient"
            fontWeight="medium"
          >
            Empowering teams to conquer project complexity.
          </Text>
          <Text className="mt-8 max-w-[600px]" color="muted" fontSize="lg">
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
          <Text className="mt-8" color="muted" fontSize="xs">
            No credit card required.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
