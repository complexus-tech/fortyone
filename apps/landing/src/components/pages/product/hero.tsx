import { Text, Button } from "ui";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="relative max-w-4xl pb-12 pt-36 text-center md:pt-40">
      <Button
        className="mx-auto px-3 text-sm md:text-base"
        color="tertiary"
        rounded="full"
        size="sm"
      >
        Our product
      </Button>
      <Text
        as="h1"
        className="my-6 pb-2 text-5xl font-semibold md:my-8 md:text-7xl"
        color="gradient"
        fontWeight="medium"
      >
        Objectives, Aligned & Achieved
      </Text>
      <Text
        className="mb-6 max-w-6xl text-xl leading-snug opacity-80 md:mb-8 md:text-2xl"
        fontWeight="normal"
      >
        We empower teams to excel with OKRs and deliver on strategic objectives.
        Our platform helps you align priorities, track progress with precision,
        and celebrate success at every milestone.
      </Text>
    </Container>
  );
};
