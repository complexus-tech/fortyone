import { Text, Button } from "ui";
import { Container } from "@/components/ui";

export const Hero = () => {
  return (
    <Container className="relative max-w-4xl pb-10 pt-36 md:pt-40">
      <Button
        className="mx-auto px-3 text-sm md:text-base"
        color="tertiary"
        rounded="full"
        size="sm"
      >
        Get in Touch
      </Button>
      <Text
        align="center"
        as="h1"
        className="my-8 pb-2 text-5xl font-semibold leading-none md:text-7xl"
        color="gradient"
      >
        How Can We Help You Today?
      </Text>
      <Text
        align="center"
        className="text-lg leading-snug opacity-80 md:mb-16 md:text-2xl"
        fontWeight="normal"
      >
        Connect with our team for product demos, implementation support, or any
        questions about getting started with complexus.
      </Text>
    </Container>
  );
};
