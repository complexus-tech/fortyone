import Link from "next/link";
import { Box, Text } from "ui";
import type { HandwrittenAccentTone } from "@/components/ui/handwritten-accent";
import { Container, UnderlinedHandwrittenAccent } from "@/components/ui";
import { SignupProviderButton } from "@/components/shared/signup-provider-button";
import { SIGNUP_URL } from "@/lib/app-url";

const HERO_TITLE =
  "Turn strategy and customer feedback into work your team can deliver.";
const HERO_DESCRIPTION =
  "Give your team a clear view of what matters, why it matters, and what to do next—so the right work keeps moving.";
const HERO_TITLE_WORDS = HERO_TITLE.split(" ");
const HERO_KEYWORD_TONES: Partial<Record<string, HandwrittenAccentTone>> = {
  strategy: "primary",
  feedback: "danger",
  work: "success",
};

export const Hero = () => {
  return (
    <Box>
      <Container className="pt-8">
        <Box className="mt-10 mb-6">
          <Text
            as="h1"
            className="relative z-1 text-5xl font-semibold text-balance md:max-w-6xl md:text-6xl"
          >
            {HERO_TITLE_WORDS.map((word, index) => {
              const keywordTone = HERO_KEYWORD_TONES[word];
              const animationStyle = {
                animationDelay: `${60 + index * 45}ms`,
              };

              return (
                <span key={word}>
                  {keywordTone ? (
                    <UnderlinedHandwrittenAccent
                      className="landing-hero-title-word"
                      style={animationStyle}
                      tone={keywordTone}
                    >
                      {word}
                    </UnderlinedHandwrittenAccent>
                  ) : (
                    <span
                      className="landing-hero-title-word inline-block"
                      style={animationStyle}
                    >
                      {word}
                    </span>
                  )}
                  {index < HERO_TITLE_WORDS.length - 1 ? " " : null}
                </span>
              );
            })}
          </Text>
          <Text className="text-text-muted mt-6 max-w-xl text-pretty">
            {HERO_DESCRIPTION}
          </Text>
        </Box>

        <Box className="landing-hero-action relative z-1 flex flex-col items-start gap-3">
          <Box className="flex flex-col items-start gap-2 sm:flex-row sm:gap-3">
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
          <Text className="ml-0.5 text-left text-[0.9rem]" color="muted">
            <Link
              className="text-foreground underline underline-offset-2"
              href={SIGNUP_URL}
            >
              Continue with email
            </Link>{" "}
            <span aria-hidden="true">&bull;</span>{" "}
            <span>No credit card required</span>
          </Text>
        </Box>
      </Container>
    </Box>
  );
};
