"use client";
import { Box } from "ui";
import { Container, Blur } from "@/components/ui";
import { useCursor } from "@/hooks";

export const ProductDemo = () => {
  const cursor = useCursor();
  return (
    <Container className="relative">
      <Box
        className="relative mx-auto aspect-[16/10] max-w-7xl overflow-hidden"
        onMouseEnter={() => {
          cursor.setText("Complexus in action");
        }}
        onMouseLeave={() => {
          cursor.removeText();
        }}
      >
        <video
          autoPlay
          className="pointer-events-none absolute -top-[4.8%] left-[10.95%] z-[2] mx-auto h-full w-[78%] object-contain opacity-90"
          loop
          muted
          playsInline
          src="/videos/demo.mp4"
        />
        <Box className="pointer-events-none absolute left-0 top-0 z-[3] h-full w-full overflow-hidden bg-[url(/images/device-mac.png)] bg-contain bg-center bg-no-repeat" />
      </Box>
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[900px] w-[900px] -translate-x-1/2 -translate-y-1/2 bg-warning/10" />
    </Container>
  );
};
