import type { ImageProps } from "next/image";
import Image from "next/image";
import { ArrowDown2Icon } from "icons";
import { cn } from "lib";
import { Box, Flex } from "ui";
import { Container, Dot } from "@/components/ui";

export type ProductScreenshotProps = {
  alt: string;
  containerClassName?: string;
  darkImage: ImageProps["src"];
  lightImage: ImageProps["src"];
  priority?: boolean;
};

export const ProductScreenshot = ({
  alt,
  containerClassName,
  darkImage,
  lightImage,
  priority = false,
}: ProductScreenshotProps) => {
  return (
    <Box data-landing-reveal>
      <Container
        className={cn("relative mt-12 overflow-visible", containerClassName)}
      >
        <Box className="relative -mr-5 w-[calc(100%+1.25rem)] overflow-hidden md:mr-0 md:w-auto md:overflow-visible">
          <Box className="border-border/70 bg-surface/90 dark:bg-surface/65 relative rounded-l-xl rounded-r-none border p-0.5 backdrop-blur md:rounded-2xl md:p-[0.35rem]">
            <Flex
              align="center"
              className="mt-1 mb-2 justify-start px-1.5 md:justify-between"
            >
              <Flex aria-hidden="true" className="gap-1.5">
                <Dot className="text-primary size-2.5" />
                <Dot className="text-warning size-2.5" />
                <Dot className="text-success size-2.5" />
              </Flex>
              <ArrowDown2Icon
                aria-hidden="true"
                className="hidden h-3.5 md:block"
                strokeWidth={2.5}
              />
            </Flex>

            <Box className="relative overflow-hidden rounded-l-lg rounded-r-none md:rounded-xl">
              <Image
                alt={alt}
                className="border-border/50 relative hidden h-88 w-auto max-w-none rounded-l-lg rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:block"
                priority={priority}
                quality={100}
                sizes="(max-width: 767px) 150vw, 100vw"
                src={darkImage}
              />
              <Image
                alt={alt}
                className="border-border/40 relative -top-px h-88 w-auto max-w-none rounded-l-lg rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:hidden"
                priority={priority}
                quality={100}
                sizes="(max-width: 767px) 150vw, 100vw"
                src={lightImage}
              />
            </Box>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
