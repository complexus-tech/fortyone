import { Button, Flex, Text, Box } from "ui";
import { ArrowRightIcon } from "icons";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[70vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 blur-3xl dark:bg-primary/5" />
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
            className="mt-6 max-w-5xl pb-2 text-4xl md:text-8xl"
            color="gradient"
          >
            Nail every objective on time.
          </Text>
          <Text
            className="mt-8 max-w-[600px] opacity-90 md:text-2xl"
            fontWeight="normal"
          >
            Empower your team to crush every key objective with our seamless
            project management platform.
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
    </Box>
  );
};
