import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const TrustTransparency = () => {
  return (
    <Container className="pb-10 pt-28">
      <Box className="mx-auto mb-20 h-1 w-16 bg-black" />
      <Box className="grid grid-cols-1 gap-8 md:grid-cols-2">
        <Box>
          <Text
            as="h2"
            className="text-6xl font-black leading-tight text-black"
          >
            Our Commitment
            <br />
            to Trust and
            <br />
            Transparency
          </Text>
          <Box className="flex items-center gap-4">
            <Text className="mt-6 text-[16rem] font-black">&ldquo;</Text>
            <Box className="flex flex-col gap-2">
              <Text className="text-3xl font-bold text-black">Giving with</Text>
              <Text className="text-3xl font-bold text-black">Confidence.</Text>
              <Text className="text-3xl font-bold text-black">
                Backing Real Impact.
              </Text>
            </Box>
          </Box>
        </Box>
        <Box>
          <Box className="flex max-w-xl flex-col gap-4">
            <Text className="text-lg">
              At AfricaGiving, we are dedicated to showcasing only credible,
              transparent, and accountable organisations. Every organisation
              featured on our platform must meet strict eligibility criteria to
              ensure they are legally registered and operate as non-profits
              within their respective countries.
            </Text>
            <Text className="text-lg">
              Our thorough verification process is designed to build confidence
              among donors, partners, and the communities these organisations
              serve. Before being listed, every organisation featured on our
              platform is assessed and given a Compliance Score out of 5, based
              on how well they meet key governance, transparency, and
              accountability standards. Each criteria met gives the organisation
              a score of 1.
            </Text>
            <Text className="text-lg">
              Once approved, organisations are added to our platform, where
              users can easily explore and support causes that align with their
              values and interests.
            </Text>
          </Box>
        </Box>
      </Box>
      <Box className="mx-auto h-1 w-32 bg-black" />
    </Container>
  );
};
