import { BlurImage, Box } from "ui";
import { Container } from "@/components/ui";

export const Content = () => {
  return (
    <Container className="py-10">
      <Box className="grid grid-cols-3 gap-10">
        <BlurImage
          alt="Transparency"
          className="aspect-square"
          src="/images/transparency/1.png"
        />
        <Box className="col-span-2">Test</Box>
      </Box>
    </Container>
  );
};
