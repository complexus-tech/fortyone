import { Box, Button, Flex, Text } from "ui";
import { Blur, Container } from "@/components/ui";

export const CallToAction = () => {
  return (
    <Box className="border-y border-gray-100 bg-gray-50 dark:border-dark-300 dark:bg-[#030303]">
      <Container className="relative max-w-7xl py-16 md:py-32">
        <Flex
          align="center"
          className="md:mt-18 mb-8 text-center"
          direction="column"
        >
          <Text
            as="h1"
            className="mt-6 h-max max-w-5xl pb-2 text-5xl md:text-7xl"
            color="gradient"
          >
            Experience the difference. Try it now!
          </Text>
          <Text
            className="mt-4 max-w-[600px] md:mt-6"
            color="muted"
            fontSize="xl"
            fontWeight="normal"
          >
            Streamline your workflows, empower your team, and nail every key
            objective all in one seamless platform.
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
        <Blur className="-translate-y-1/23 absolute -bottom-20 left-1/2 right-1/2 top-1/2 h-[800px] w-[800px] -translate-x-1/2 bg-warning/5" />
      </Container>
    </Box>
  );
};
