import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

const HERO_TITLE =
  "Turn strategy and customer feedback into work your team can deliver.";
const HERO_TITLE_WORDS = HERO_TITLE.split(" ");

export const Hero = () => {
  return (
    <Box>
      <Container className="pt-12">
        <Box className="mt-12 mb-6 md:mt-24">
          <Text
            as="h1"
            className="relative z-1 text-5xl font-medium text-balance md:max-w-6xl md:text-6xl"
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
