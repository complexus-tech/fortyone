import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container>
      <Box className="bg-secondary py-16">
        <Box className="mx-auto max-w-4xl px-4">
          <Text className="text-center text-7xl font-bold text-white">
            Submit a Story
          </Text>
        </Box>
      </Box>
    </Container>
  );
};
