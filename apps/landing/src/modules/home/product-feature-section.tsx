import type { ReactNode } from "react";
import { Box, Text } from "ui";
import {
  NavigationMenuIcon,
  type NavigationIconName,
  type NavigationIconTone,
} from "@/components/shared/navigation-menu-icon";
import { Container } from "@/components/ui";

type ProductFeatureSectionProps = {
  children: ReactNode;
  description: string;
  eyebrow?: string;
  id: string;
  icon?: NavigationIconName;
  iconTone?: NavigationIconTone;
  title: ReactNode;
};

export const ProductFeatureSection = ({
  children,
  description,
  eyebrow,
  id,
  icon,
  iconTone = "blue",
  title,
}: ProductFeatureSectionProps) => {
  return (
    <Box as="section" className="scroll-mt-20 py-16 md:py-24" id={id}>
      <Container>
        <Box className="border-border/70 border-t-[0.5px] pt-12 md:pt-16">
          <Box className="flex flex-col gap-6 md:flex-row md:items-end md:justify-between md:gap-16">
            <Box data-landing-reveal>
              {icon ? (
                <Box className="mb-5 flex items-center gap-3">
                  <NavigationMenuIcon
                    className="size-11"
                    name={icon}
                    tone={iconTone}
                  />
                  {eyebrow ? (
                    <Text className="text-text-muted text-xs font-medium tracking-[0.11em] uppercase">
                      {eyebrow}
                    </Text>
                  ) : null}
                </Box>
              ) : null}
              <Text
                as="h2"
                className="max-w-4xl pb-1 text-4xl text-balance md:text-5xl"
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
