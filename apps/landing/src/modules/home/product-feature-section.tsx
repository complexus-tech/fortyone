import type { ReactNode } from "react";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";

type ProductFeatureSectionProps = {
  children: ReactNode;
  compact?: boolean;
  description: string;
  eyebrow?: string;
  id: string;
  title: ReactNode;
};

export const ProductFeatureSection = ({
  children,
  compact = false,
  description,
  eyebrow,
  id,
  title,
}: ProductFeatureSectionProps) => {
  return (
    <Box
      as="section"
      className={cn(
        "scroll-mt-20",
        compact ? "py-10 md:py-14" : "py-16 md:py-24",
      )}
      id={id}
    >
      <Container>
        <Box
          className={cn(
            "border-border/70 border-t-[0.5px]",
            compact ? "pt-8 md:pt-10" : "pt-12 md:pt-16",
          )}
        >
          <Box className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between md:gap-16">
            <Box data-landing-reveal>
              {eyebrow ? (
                <Text className="text-primary mb-4 text-xs font-semibold tracking-[0.16em] uppercase">
                  {eyebrow}
                </Text>
              ) : null}
              <Text
                as="h2"
                className={cn(
                  "max-w-4xl pb-1 text-balance",
                  compact ? "text-3xl md:text-4xl" : "text-4xl md:text-5xl",
                )}
              >
                {title}
              </Text>
            </Box>
            <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
              <Text className="w-full max-w-xl text-base leading-relaxed opacity-70 md:mb-0.5">
                {description}
              </Text>
            </Box>
          </Box>
        </Box>
      </Container>

      {children}
    </Box>
  );
};
