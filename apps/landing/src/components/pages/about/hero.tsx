import { Flex, Text, Box } from "ui";
import { Container, Blur } from "@/components/ui";

export const Hero = () => {
  return (
    <Box>
      <Blur className="absolute -top-[65vh] left-1/2 right-1/2 h-screen w-screen -translate-x-1/2 bg-primary/15 dark:bg-primary/[0.03]" />
      <Container className="max-w-5xl pt-20">
        <Flex
          align="center"
          className="mb-8 mt-20 text-center"
          direction="column"
        >
          <Text
            as="h1"
            className="mt-6 h-max max-w-5xl pb-2 text-4xl md:text-7xl"
            color="gradient"
            fontWeight="medium"
          >
            Project Management liberated as open source.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
