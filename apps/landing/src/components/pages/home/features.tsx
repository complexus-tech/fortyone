import { Box, Text, Flex, Wrapper } from "ui";
import { StoryIcon } from "icons";
import Image from "next/image";
import { Container, Blur } from "@/components/ui";

const Intro = () => (
  <Box className="relative">
    <Box as="section" className="my-20 text-center">
      <Text
        as="h3"
        className="mx-auto max-w-4xl pb-2 text-4xl font-medium md:text-7xl"
        color="gradient"
      >
        Say goodbye to wasted time and energy.
      </Text>
      <Text className="mx-auto mt-2 max-w-[700px] md:mt-6" fontSize="lg">
        Simplify workflows, streamline collaboration, and achieve exceptional
        results with Complexus. With features like OKR Tracking, Themes
        Management, Iterations Planning, and Roadmap Visualization, welcome to
        effortless project management.
      </Text>
    </Box>
    <Blur className="absolute left-1/2 right-1/2 top-28 h-[900px] w-[900px] -translate-x-1/2 bg-primary/40 dark:bg-secondary/20" />
  </Box>
);

export const Features = () => {
  return (
    <Container as="section">
      <Intro />
      <Box className="grid grid-cols-3 gap-8">
        <Wrapper className="rounded-3xl px-8 py-10 shadow-2xl dark:bg-dark-300/30">
          <Flex align="center" className="mb-6" gap={4} justify="between">
            <Text as="h3" fontSize="2xl" fontWeight="medium">
              Stories
            </Text>
            <StoryIcon className="h-8 w-auto text-primary" />
          </Flex>
          <Text className="mb-8" color="muted">
            Lorem ipsum dolor sit amet, consectetur adipisicing elit. Placeat
            cupiditate nam quis, iure esse magnam.
          </Text>
          <Image
            alt="Stories"
            className="mx-auto block rounded-lg"
            height={1134}
            src="/sto.png"
            width={752}
          />
        </Wrapper>
        <Wrapper className="min-h-[50vh] rounded-3xl px-8 py-10 shadow-2xl dark:bg-dark-300/30">
          Test
        </Wrapper>
        <Wrapper className="min-h-[50vh] rounded-3xl px-8 py-10 shadow-2xl dark:bg-dark-300/30">
          Test
        </Wrapper>
        <Wrapper className="col-span-2 min-h-[40vh] rounded-3xl px-8 py-10 shadow-2xl dark:bg-dark-300/30">
          Test
        </Wrapper>
        <Wrapper className="min-h-[40vh] rounded-3xl px-8 py-10 shadow-2xl dark:bg-dark-300/30">
          Test
        </Wrapper>
      </Box>
    </Container>
  );
};
