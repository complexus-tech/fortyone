import { Text } from "ui";
import { Container } from "./container";

export const ComingSoon = () => {
  return (
    <Container className="relative max-w-4xl pb-16 pt-36 md:pt-40">
      <Text
        align="center"
        as="h1"
        className="mb-8 text-5xl leading-[1.05] md:text-7xl"
        color="gradient"
        fontWeight="medium"
      >
        Coming Soon
      </Text>
      <Text
        align="center"
        className="text-lg leading-snug opacity-80 md:text-2xl"
        fontWeight="normal"
      >
        We&apos;re working hard to bring you something amazing. Stay tuned!
      </Text>
    </Container>
  );
};
