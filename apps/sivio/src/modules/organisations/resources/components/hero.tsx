import { Box, Text, BlurImage, Flex } from "ui";

export const Hero = () => {
  return (
    <Box className="relative">
      <BlurImage
        className="aspect-[16/6]"
        quality={100}
        src="/images/about/4.jpg"
      />
      <Flex
        align="center"
        className="absolute inset-0 bg-black/50"
        direction="column"
        justify="center"
      >
        <Text className="mb-10 text-center text-8xl font-semibold text-white">
          Recommended <br /> Resources
        </Text>
        <Text className="text-center text-3xl text-white">
          &ldquo;Your Giving, Backed by Insight.&rdquo;
        </Text>
      </Flex>
    </Box>
  );
};
