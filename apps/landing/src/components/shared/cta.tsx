import Image from "next/image";
import { Box, Button } from "ui";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";
import meshImage from "../../../public/images/meshing.webp";

export const CallToAction = () => {
  return (
    <Container className="py-16 md:py-20">
      <Box data-landing-reveal>
        <Box className="relative flex flex-col items-center justify-center overflow-hidden rounded-2xl md:rounded-3xl">
          <Image
            alt=""
            className="object-cover"
            fill
            quality={100}
            sizes="100vw"
            src={meshImage}
          />
          <Box className="absolute inset-0 z-1 dark:bg-black/30" />
          <Box className="relative z-2 flex max-w-[760px] flex-col items-center px-6 py-24 md:py-36">
            <h2 className="text-foreground text-center text-4xl font-medium tracking-tight md:text-6xl">
              Turn customer feedback into work your team can deliver.
            </h2>
            <Button
              className="mt-8 border-0"
              color="invert"
              href={SIGNUP_URL}
              rounded="lg"
              size="lg"
            >
              Start free
            </Button>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
