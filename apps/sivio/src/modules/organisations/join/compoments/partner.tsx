import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Partner = () => {
  return (
    <Container className="pt-16">
      <Box className="mx-auto max-w-6xl">
        <Box className="mb-16 grid grid-cols-1 gap-12 md:grid-cols-5">
          <Text className="text-6xl font-bold md:col-span-2">
            Partner with AfricaGiving — Amplify Your Impact Across the Continent
          </Text>
          <Box className="md:col-span-3">
            <Text className="mb-10 text-lg">
              AfricaGiving is a pan-African fundraising platform that supports
              nonprofits, community-based organizations, and social impact
              initiatives in mobilizing resources across Africa and around the
              world.
            </Text>
            <Text className="text-lg">
              We offer a trusted, secure, and user-friendly space for African
              organisations to tell their story, connect with donors, and
              receive donations seamlessly
            </Text>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
