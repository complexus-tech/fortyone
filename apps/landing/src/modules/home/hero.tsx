import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

const HERO_TITLE = "Plan and deliver projects with customer feedback built in.";
const HERO_TITLE_WORDS = HERO_TITLE.split(" ");

export const Hero = () => {
  return (
    <Box>
      <Container className="pt-12">
        <Box className="mt-12 mb-6 flex flex-col gap-6 md:mt-24 md:flex-row md:items-end md:justify-between md:gap-12">
          <Text
            as="h1"
            className="relative z-1 text-5xl font-medium text-balance md:max-w-7xl md:text-6xl"
          >
            {HERO_TITLE_WORDS.map((word, index) => (
              <span key={`${word}-${index}`}>
                <span
                  className="landing-hero-title-word inline-block"
                  style={{ animationDelay: `${60 + index * 45}ms` }}
                >
                  {word}
                </span>
                {index < HERO_TITLE_WORDS.length - 1 ? " " : null}
              </span>
            ))}
          </Text>

          <Box className="landing-hero-copy">
            <Text className="w-full max-w-xl opacity-60 md:mb-0.5">
              Collect requests, decide what matters, and move accepted feedback
              into the project plan without losing its context.
            </Text>
          </Box>
        </Box>

        <Box className="landing-hero-action">
          <Button
            className="relative z-1 px-3 md:pr-4 md:pl-5"
            color="invert"
            href={SIGNUP_URL}
            rounded="lg"
            size="lg"
          >
            Get started free
          </Button>
        </Box>
      </Container>
    </Box>
  );
};
