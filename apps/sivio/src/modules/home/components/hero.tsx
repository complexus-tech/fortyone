import { BlurImage, Box, Flex, Text } from "ui";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="mt-4">
      <Box className="relative">
        <BlurImage
          className="aspect-video"
          quality={100}
          src="/images/home/hero.webp"
        />
        <Box className="absolute inset-0 grid grid-cols-2 bg-black/20">
          Test
          <Flex
            className="bg-secondary/80"
            direction="column"
            justify="between"
          >
            <Box className="p-10">
              <Text as="h1" className="mb-8 text-7xl font-semibold text-white">
                Be part of the change. <br /> Give today.
              </Text>
              <Text className="max-w-xs text-lg text-white">
                Your donation can help transform lives and bring hope to
                communities.
              </Text>
            </Box>
            <Box className="grid grid-cols-3">
              <Box>
                <Text>Education</Text>
              </Box>
            </Box>
          </Flex>
        </Box>
      </Box>
    </Container>
  );
};
