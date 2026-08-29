import type { ImageProps } from "next/image";
import { Box, Text } from "ui";
import { SignupProviderButton } from "@/components/shared/signup-provider-button";
import { Container } from "@/components/ui";
import { ProductScreenshot } from "@/modules/home/product-screenshot";

export type FeatureDetailHeroProps = {
  description: string;
  imageAlt: string;
  imageDark: ImageProps["src"];
  imageLight: ImageProps["src"];
  title: string;
  url: string;
};

export function FeatureDetailHero({
  description,
  imageAlt,
  imageDark,
  imageLight,
  title,
  url,
}: FeatureDetailHeroProps) {
  return (
    <section
      aria-labelledby="feature-detail-hero-title"
      className="landing-hero-shell landing-page-frame mt-18 rounded-2xl pt-px pb-6 sm:rounded-[3rem] md:mt-20 md:rounded-[4rem] md:pb-10"
    >
      <header>
        <Container className="pt-8">
          <Box className="mt-10 mb-6">
            <Text
              as="h1"
              className="relative z-1 max-w-2xl text-5xl font-semibold text-balance md:text-6xl"
              id="feature-detail-hero-title"
            >
              {title}
            </Text>
            <Text className="text-text-muted mt-6 max-w-xl text-pretty">
              {description}
            </Text>
          </Box>

          <Box className="landing-hero-action relative z-1 flex flex-col items-start gap-2 sm:flex-row sm:gap-3">
            <SignupProviderButton
              className="px-3 md:px-3"
              emphasized
              label="Continue"
              provider="google"
            />
            <SignupProviderButton
              className="px-3 md:px-3"
              emphasized
              label="Continue"
              provider="microsoft"
            />
          </Box>
        </Container>
      </header>

      <ProductScreenshot
        alt={imageAlt}
        cropBrowserOnMobile
        darkImage={imageDark}
        lightImage={imageLight}
        priority
        url={url}
      />
    </section>
  );
}
