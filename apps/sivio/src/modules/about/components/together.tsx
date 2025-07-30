import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Together = () => {
  return (
    <Container className="py-28">
      <Box className="grid grid-cols-1 gap-12 md:grid-cols-3">
        <Box>
          <Text
            as="h2"
            className="text-6xl font-black leading-tight text-black"
          >
            Together, we are building a culture of giving back by Africans, for
            Africa.
          </Text>
        </Box>
        <Box className="md:col-span-2">
          <Box className="flex max-w-xl flex-col gap-4">
            <Text className="text-lg">
              AfricaGiving is a platform that connects everyday givers to
              nonprofit organisations working to create positive change across
              Africa. It provides a trusted space for individuals, businesses,
              and communities to support local causes and contribute to
              sustainable development on the continent.
            </Text>
            <Text className="text-lg">
              AfricaGiving is an initiative of the SIVIO Institute, an
              independent organisation dedicated to strengthening citizen agency
              and advancing inclusive, evidence-based development. Through
              AfricaGiving, SIVIO Institute promotes local philanthropy and
              greater accountability by empowering African nonprofits to tell
              their stories, raise resources, and build lasting relationships
              with their supporters.
            </Text>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
