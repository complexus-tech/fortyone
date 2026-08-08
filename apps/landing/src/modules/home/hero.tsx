import { Box, Button, Text } from "ui";
import { cn } from "lib";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";

const HERO_TITLE =
  "Turn strategy and customer feedback into work your team can deliver.";
const HERO_TITLE_WORDS = HERO_TITLE.split(" ");
const HERO_KEYWORD_COLOR_CLASS_NAMES: Partial<Record<string, string>> = {
  strategy: "text-primary",
  feedback: "text-danger",
  work: "text-success",
};

const HeroKeywordUnderline = () => {
  return (
    <svg
      aria-hidden="true"
      className="pointer-events-none absolute -bottom-[0.02em] left-0 h-[0.16em] w-full overflow-visible"
      fill="none"
      preserveAspectRatio="none"
      viewBox="0 0 100 12"
    >
      <path
        d="M2 7.8C12 5.2 19 8.6 30 6.5C42 4.2 51 8.4 62 6.1C74 3.8 86 7.9 98 5.2"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="2.4"
        vectorEffect="non-scaling-stroke"
      />
      <path
        d="M5 10.1C22 8.4 38 10.4 55 8.3C70 6.5 84 9.4 96 7.6"
        opacity="0.38"
        stroke="currentColor"
        strokeLinecap="round"
        strokeWidth="1.2"
        vectorEffect="non-scaling-stroke"
      />
    </svg>
  );
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
              const keywordColorClassName =
                HERO_KEYWORD_COLOR_CLASS_NAMES[word];

              return (
                <span key={word}>
                  <span
                    className={cn(
                      "landing-hero-title-word inline-block",
                      keywordColorClassName
                        ? "font-handwritten relative text-[1.08em] leading-[0.9] font-bold tracking-[-0.045em]"
                        : undefined,
                      keywordColorClassName,
                    )}
                    style={{ animationDelay: `${60 + index * 45}ms` }}
                  >
                    {word}
                    {keywordColorClassName ? <HeroKeywordUnderline /> : null}
                  </span>
                  {index < HERO_TITLE_WORDS.length - 1 ? " " : null}
                </span>
              );
            })}
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
