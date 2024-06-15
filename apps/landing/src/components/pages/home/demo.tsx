"use client";
import { Box } from "ui";
import { useState } from "react";
import { motion } from "framer-motion";
import { Container, Blur } from "@/components/ui";
import { useCursor } from "@/hooks";

export const ProductDemo = () => {
  const [isActive, setIsActive] = useState(false);
  const cursor = useCursor();
  return (
    <Container className="relative">
      <Box
        onMouseEnter={() => {
          cursor.setText("See in action");
          setIsActive(true);
        }}
        onMouseLeave={() => {
          cursor.removeText();
          setIsActive(false);
        }}
      >
        <motion.div
          animate={isActive ? { y: -6, x: 6 } : { y: 0, x: 0 }}
          className="relative z-[2] mx-auto aspect-[16/10] max-w-7xl overflow-hidden"
          initial={{ y: 0, x: 0 }}
          transition={{ type: "spring", stiffness: 100 }}
        >
          <video
            autoPlay
            className="pointer-events-none absolute -top-[4.6%] left-[10.95%] z-[2] mx-auto h-full w-[78%] object-contain opacity-90"
            loop
            muted
            playsInline
            src="/videos/demo.mp4"
          />
          <Box className="pointer-events-none absolute left-0 top-0 z-[3] h-full w-full overflow-hidden bg-[url(/images/device-mac.png)] bg-contain bg-center bg-no-repeat" />
        </motion.div>
      </Box>
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 z-0 h-[300px] w-[300px] -translate-x-1/2 -translate-y-1/2 bg-warning/10 md:h-[900px] md:w-[900px]" />
    </Container>
  );
};
