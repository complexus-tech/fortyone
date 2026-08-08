import { Box, Button, Text } from "ui";
import type { HandwrittenAccentTone } from "@/components/ui/handwritten-accent";
import { Container, UnderlinedHandwrittenAccent } from "@/components/ui";
import { GoogleSignupButton } from "@/components/shared/google-signup-button";
import { SIGNUP_URL } from "@/lib/app-url";

const HERO_TITLE =
  "Turn strategy and customer feedback into work your team can deliver.";
const HERO_TITLE_WORDS = HERO_TITLE.split(" ");
const HERO_KEYWORD_TONES: Partial<Record<string, HandwrittenAccentTone>> = {
  strategy: "primary",
  feedback: "danger",
  work: "success",
};

export const Hero = () => {
  return (
    <Box>
      <Container className="pt-12">
        <Box className="mt-12 mb-6 md:mt-24">
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
          <Text className="sr-only">
            FortyOne is an AI project management platform that connects company
            strategy and customer feedback to objectives, plans, and daily work,
            so teams can decide what matters and deliver it with context.
          </Text>
        </Box>

        <Box className="landing-hero-action flex flex-col items-start gap-2 sm:flex-row sm:items-center sm:gap-3">
          <Button
            className="relative z-1 px-3 md:pr-4 md:pl-5"
            color="invert"
            href={SIGNUP_URL}
            rounded="lg"
            size="lg"
          >
            Start for free
          </Button>
          <GoogleSignupButton />
        </Box>
      </Container>
    </Box>
  );
};
