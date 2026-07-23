import type { ImageProps } from "next/image";
import Image from "next/image";
import { ArrowDown2Icon, ArrowLeft2Icon, RefreshIcon } from "icons";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import { Container, Dot } from "@/components/ui";

export type ProductScreenshotProps = {
  alt: string;
  containerClassName?: string;
  cropBrowserOnMobile?: boolean;
  darkImage: ImageProps["src"];
  lightImage: ImageProps["src"];
  priority?: boolean;
  url: string;
};

export const ProductScreenshot = ({
  alt,
  containerClassName,
  cropBrowserOnMobile = false,
  darkImage,
  lightImage,
  priority = false,
  url,
}: ProductScreenshotProps) => {
  return (
    <Box data-landing-reveal>
      <Container
        className={cn("relative mt-12 overflow-visible", containerClassName)}
      >
        <Box className="relative -mr-5 w-[calc(100%+1.25rem)] overflow-hidden md:mr-0 md:w-auto md:overflow-visible">
          <Box
            className={cn(
              "border-border/70 bg-surface/90 dark:bg-surface/65 relative rounded-l-xl rounded-r-none border p-0.5 backdrop-blur-md md:rounded-2xl md:p-[0.35rem]",
              {
                "border-r-0 pr-0 md:border-r md:pr-[0.35rem]":
                  cropBrowserOnMobile,
              },
            )}
          >
            <Flex
              align="center"
              className={cn(
                "relative justify-start px-1.5 md:mt-1 md:mb-2 md:min-h-4 md:justify-between",
                cropBrowserOnMobile
                  ? "mt-0.5 mb-1 min-h-3"
                  : "mt-1 mb-2 min-h-4",
              )}
            >
              <Flex aria-hidden="true" className="gap-1.5">
                <Dot className="text-primary size-2.5" />
                <Dot className="text-warning size-2.5" />
                <Dot className="text-success size-2.5" />
              </Flex>
              <Flex
                align="center"
                className={cn(
                  "absolute left-1/2 -translate-x-1/2 gap-2",
                  cropBrowserOnMobile && "hidden md:flex",
                )}
              >
                <ArrowLeft2Icon
                  aria-hidden="true"
                  className="text-text-muted hidden h-3.5 opacity-70 sm:block"
                  strokeWidth={2.25}
                />
                <Text
                  as="span"
                  className="bg-surface-muted text-text-muted dark:bg-surface-elevated max-w-[calc(100vw-7rem)] truncate rounded-md px-2 py-0.5 text-[0.625rem] leading-4 font-medium md:max-w-md md:text-center md:text-xs"
                >
                  {url}
                </Text>
                <RefreshIcon
                  aria-hidden="true"
                  className="text-text-muted hidden h-3.5 opacity-70 sm:block"
                  strokeWidth={2.25}
                />
              </Flex>
              <ArrowDown2Icon
                aria-hidden="true"
                className="text-text-muted hidden h-3.5 opacity-70 md:block"
                strokeWidth={2.25}
              />
            </Flex>

            <Box className="relative overflow-hidden rounded-l-lg rounded-r-none md:rounded-xl">
              <Image
                alt={alt}
                className={cn(
                  "border-border/50 relative hidden h-88 w-auto max-w-none rounded-l-lg rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:block",
                  cropBrowserOnMobile && "border-r-0 md:border-r",
                )}
                priority={priority}
                quality={100}
                sizes="(max-width: 767px) 150vw, 100vw"
                src={darkImage}
              />
              <Image
                alt={alt}
                className={cn(
                  "border-border/40 relative h-88 w-auto max-w-none rounded-l-lg rounded-r-none border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:hidden",
                  cropBrowserOnMobile && "border-r-0 md:border-r",
                )}
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
