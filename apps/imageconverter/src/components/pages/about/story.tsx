import { Text } from "ui";
import { Container } from "@/components/ui";

export const Story = () => {
  return (
    <Container className="max-w-3xl">
      <Text
        className="mb-4 text-xl uppercase tracking-wide md:mb-6 md:md:w-1/2 md:text-2xl"
        color="gradient"
      >
        Our Mission
      </Text>
      <Text
        className="mb-20 max-w-6xl text-xl leading-snug opacity-80 md:mb-32 md:text-2xl"
        fontWeight="normal"
      >
        Our mission is to provide a seamless, efficient, and accessible online
        image conversion solution. We strive to empower users worldwide with the
        ability to transform their images effortlessly, supporting creativity
        and productivity across various platforms and devices.
      </Text>
    </Container>
  );
};
