import { Flex, Text, Box } from "ui";
import { Blur } from "./blur";
import { Container } from "./container";

export const ComingSoon = () => {
  return (
    <Box className="mb-40">
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/3" />
      <Container className="pt-20">
        <Flex
          align="center"
          className="mb-8 mt-20 text-center"
          direction="column"
        >
          <Text
            as="h1"
            className="mb-16 mt-6 h-max max-w-5xl pb-2 text-4xl md:text-8xl"
            color="gradient"
            fontWeight="medium"
          >
            Coming Soon.
          </Text>

          <Text color="muted" fontSize="2xl">
            We are working on something awesome. Stay tuned!
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
