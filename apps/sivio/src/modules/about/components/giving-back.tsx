import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const GivingBack = () => {
  return (
    <Container className="py-10">
      <Box className="grid grid-cols-1 gap-8 md:grid-cols-3">
        <Box>
          <Box className="mb-4 h-2 w-36 bg-secondary" />
          <Text
            as="h2"
            className="text-6xl font-black leading-tight text-black"
          >
            Giving
            <br />
            Back, The
            <br />
            African Way
          </Text>
        </Box>
        <Box className="md:col-span-2">
          <Box className="flex max-w-2xl flex-col gap-4">
            <Text className="text-lg">
              AfricaGiving is a platform dedicated to unlocking the power of
              giving across Africa. Our core purpose is to connect donors,{" "}
              <span className="font-bold">individuals</span>,{" "}
              <span className="font-bold">corporates</span>, and{" "}
              <span className="font-bold">foundations</span>, with credible,
              high-impact African-led organisations that are changing lives
              across the continent.
            </Text>
            <Text className="text-lg">
              We believe that Africans, and friends of Africa, have the will,
              resources and networks to drive transformative change from within.
              That&apos;s why AfricaGiving exists to make giving{" "}
              <span className="font-bold">easier</span>, more{" "}
              <span className="font-bold">transparent</span>, and{" "}
              <span className="font-bold">deeply</span> connected to local
              realities.
            </Text>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
