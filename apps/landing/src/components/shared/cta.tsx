"use client";
import Image from "next/image";
import { Box, Button } from "ui";
import { motion } from "framer-motion";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";
import meshImage from "../../../public/images/meshing.webp";

const viewport = { once: true, amount: 0.35 };
const scaleIn = {
  hidden: { scale: 0.98, opacity: 0 },
  show: {
    scale: 1,
    opacity: 1,
    transition: { duration: 0.7, ease: "easeOut" },
  },
};

export const CallToAction = () => {
  return (
    <Container className="py-16 md:py-20">
      <motion.div
        initial="hidden"
        variants={scaleIn}
        viewport={viewport}
        whileInView="show"
      >
        <Box className="relative flex flex-col items-center justify-center overflow-hidden rounded-2xl md:rounded-3xl">
          <Image
            alt=""
            className="object-cover"
            src={meshImage}
            fill
            quality={100}
          />
          <Box className="absolute inset-0 z-1 dark:bg-black/30" />
          <Box className="relative z-2 flex flex-col items-center px-6 py-24 md:py-36">
            <h2 className="text-foreground text-center text-4xl font-medium tracking-tight md:text-6xl">
              Try FortyOne today.
            </h2>
            <Button
              className="mt-8 border-0"
              color="invert"
              href={SIGNUP_URL}
              rounded="lg"
              size="lg"
            >
              Get started
            </Button>
          </Box>
        </Box>
      </motion.div>
    </Container>
  );
};
