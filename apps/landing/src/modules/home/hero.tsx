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
const HERO_KEYWORD_INK_DROP_CLASS_NAMES: Partial<
  Record<string, readonly string[]>
> = {
  strategy: [
    "-top-[0.11em] left-[20%] size-[0.045em] opacity-55",
    "-top-[0.04em] -right-[0.1em] h-[0.085em] w-[0.05em] rotate-[18deg] opacity-65",
    "bottom-[0.02em] -left-[0.07em] size-[0.045em] opacity-50",
    "-top-[0.1em] right-[17%] size-[0.03em] opacity-45",
    "-bottom-[0.01em] right-[3%] size-[0.025em] opacity-40",
  ],
  feedback: [
    "-top-[0.08em] left-[7%] h-[0.05em] w-[0.075em] -rotate-12 opacity-50",
    "top-[0.1em] -right-[0.1em] size-[0.065em] opacity-65",
    "-bottom-[0.01em] left-[52%] size-[0.035em] opacity-45",
    "bottom-[0.03em] -left-[0.08em] h-[0.08em] w-[0.045em] rotate-12 opacity-65",
    "-top-[0.12em] right-[22%] size-[0.028em] opacity-45",
  ],
  work: [
    "-top-[0.1em] left-[30%] size-[0.055em] opacity-55",
    "top-[0.12em] -right-[0.11em] h-[0.045em] w-[0.075em] rotate-[20deg] opacity-55",
    "bottom-[0.02em] -left-[0.08em] size-[0.055em] opacity-65",
    "-top-[0.04em] -left-[0.08em] size-[0.03em] opacity-45",
    "-bottom-[0.02em] right-[18%] size-[0.03em] opacity-45",
  ],
};

const HeroKeywordInkDrops = ({ word }: { word: string }) => {
  const dropClassNames = HERO_KEYWORD_INK_DROP_CLASS_NAMES[word];

  if (!dropClassNames) return null;

  return (
    <span
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 mix-blend-multiply filter-[brightness(0.88)] dark:mix-blend-normal dark:filter-[brightness(0.68)]"
    >
      {dropClassNames.map((className) => (
        <span
          className={cn(
            "absolute rounded-[55%_45%_60%_40%] bg-current",
            className,
          )}
          key={className}
        />
      ))}
    </span>
  );
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
                        ? "font-handwritten relative text-[1.08em] leading-[0.9] tracking-[-0.045em]"
                        : undefined,
                      keywordColorClassName,
                    )}
                    style={{ animationDelay: `${60 + index * 45}ms` }}
                  >
                    {word}
                    {keywordColorClassName ? (
                      <HeroKeywordInkDrops word={word} />
                    ) : null}
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
