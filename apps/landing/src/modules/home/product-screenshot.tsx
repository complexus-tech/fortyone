import type { ImageProps } from "next/image";
import Image from "next/image";
import { ArrowDown2Icon, ArrowLeft2Icon, RefreshIcon } from "icons";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import { Container, Dot } from "@/components/ui";
import styles from "./product-screenshot.module.css";

export type ProductScreenshotProps = {
  alt: string;
  containerClassName?: string;
  cropBrowserOnMobile?: boolean;
  darkImage: ImageProps["src"];
  lightImage: ImageProps["src"];
  priority?: boolean;
  reveal?: boolean;
  url: string;
};

export const ProductScreenshot = ({
  alt,
  containerClassName,
  cropBrowserOnMobile = false,
  darkImage,
  lightImage,
  priority = false,
  reveal = true,
  url,
}: ProductScreenshotProps) => {
  return (
    <Box data-landing-reveal={reveal ? "" : undefined}>
      <Container
        className={cn("relative mt-12 overflow-visible", containerClassName)}
      >
        <Box
          className={cn(
            "relative md:mr-0 md:w-auto md:overflow-visible",
            cropBrowserOnMobile
              ? "-mr-5 w-[calc(100%+1.25rem)] overflow-hidden"
              : "w-full overflow-visible",
          )}
        >
          <Box
            className={cn(
              "bg-surface/90 dark:bg-surface shadow-border/70 relative border border-transparent p-0.5 shadow-2xl backdrop-blur-md md:rounded-2xl md:border-r md:p-[0.35rem] dark:shadow-none",
              styles.frame,
              cropBrowserOnMobile
                ? "rounded-l-xl rounded-r-none border-r-0 pr-0"
                : "rounded-xl",
            )}
          >
            <Flex
              align="center"
              className={cn(
                "relative mt-0.5 mb-1 min-h-3 justify-start px-2 md:mt-1 md:mb-2 md:min-h-4 md:justify-between",
                styles.chrome,
              )}
            >
              <Flex aria-hidden="true" className="gap-1.5">
                <Dot className="text-primary size-2.5" />
                <Dot className="text-warning size-2.5" />
                <Dot className="text-success size-2.5" />
              </Flex>
              <Flex
                align="center"
                className="absolute left-1/2 hidden -translate-x-1/2 gap-2 md:flex"
              >
                <ArrowLeft2Icon
                  aria-hidden="true"
                  className="text-text-muted hidden h-3.5 opacity-80 sm:block"
                  strokeWidth={2.25}
                />
                <Text
                  as="span"
                  className="bg-surface-muted/80 text-text-muted max-w-[calc(100vw-7rem)] truncate rounded-xl px-2 py-0.5 text-[0.625rem] leading-4 font-medium backdrop-blur-sm md:max-w-md md:text-center md:text-xs dark:bg-white/10"
                >
                  {url}
                </Text>
                <RefreshIcon
                  aria-hidden="true"
                  className="text-text-muted hidden h-3.5 opacity-80 sm:block"
                  strokeWidth={2.25}
                />
              </Flex>
              <ArrowDown2Icon
                aria-hidden="true"
                className="text-text-muted hidden h-3.5 opacity-80 md:block"
                strokeWidth={2.25}
              />
            </Flex>

            <Box
              className={cn(
                "relative overflow-hidden md:rounded-xl",
                cropBrowserOnMobile
                  ? "rounded-l-lg rounded-r-none"
                  : "rounded-lg",
              )}
            >
              <Image
                alt={alt}
                className={cn(
                  "border-border/50 dark:border-border/30 relative hidden border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:block",
                  cropBrowserOnMobile
                    ? "h-96 w-auto max-w-none rounded-l-lg rounded-r-none border-r-0 md:border-r"
                    : "h-auto w-full max-w-full rounded-lg",
                )}
                priority={priority}
                sizes={
                  cropBrowserOnMobile
                    ? "(max-width: 767px) 150vw, 100vw"
                    : "100vw"
                }
                src={darkImage}
              />
              <Image
                alt={alt}
                className={cn(
                  "border-border/40 relative border md:h-auto md:w-full md:max-w-full md:rounded-xl dark:hidden",
                  cropBrowserOnMobile
                    ? "h-96 w-auto max-w-none rounded-l-lg rounded-r-none border-r-0 md:border-r"
                    : "h-auto w-full max-w-full rounded-lg",
                )}
                priority={priority}
                sizes={
                  cropBrowserOnMobile
                    ? "(max-width: 767px) 150vw, 100vw"
                    : "100vw"
                }
                src={lightImage}
              />
            </Box>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
