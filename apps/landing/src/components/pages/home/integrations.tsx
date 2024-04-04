import { Flex, Text, Box } from "ui";
import { Container } from "@/components/ui";

export const Integrations = () => {
  return (
    <Box className="bg-dark py-40">
      <Container>
        <Flex
          align="center"
          className="md:mt-18 mb-8 text-center"
          direction="column"
        >
          <Text
            as="h1"
            className="mt-6 h-max max-w-4xl pb-2 text-7xl"
            color="gradient"
            fontWeight="medium"
          >
            Sync up your favorite tools.
          </Text>
          <Text
            className="mt-4 max-w-[600px] md:mt-6"
            color="muted"
            fontSize="lg"
          >
            Complexus seamlessly integrates with popular software tools like
            GitHub, Miro, Figma, Notion, Slack, and Loom, enhancing project
            management efficiency.
          </Text>
        </Flex>
      </Container>
    </Box>
  );
};
