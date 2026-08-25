import Image from "next/image";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";

type Integration = {
  detail: string;
  logoClassName?: string;
  name: string;
  src: string;
};

const integrations: readonly Integration[] = [
  { detail: "Requests", name: "Slack", src: "/integrations/slack.svg" },
  {
    detail: "Design context",
    name: "Figma",
    src: "/integrations/figma.svg",
  },
  {
    detail: "Delivery",
    logoClassName: "dark:invert",
    name: "GitHub",
    src: "/integrations/github-mark.svg",
  },
  {
    detail: "Availability",
    name: "Google Calendar",
    src: "/integrations/google-calendar-2026.svg",
  },
  {
    detail: "Availability",
    name: "Outlook Calendar",
    src: "/integrations/outlook-2025.svg",
  },
  {
    detail: "Reply to Maya",
    name: "Gmail",
    src: "/integrations/gmail-2020.png",
  },
  {
    detail: "MCP",
    logoClassName: "dark:invert",
    name: "ChatGPT",
    src: "/integrations/chatgpt-mark.svg",
  },
  {
    detail: "MCP",
    name: "Claude",
    src: "/integrations/claude-color.svg",
  },
  {
    detail: "MCP",
    logoClassName: "dark:invert",
    name: "Cursor",
    src: "/integrations/cursor.svg",
  },
];

export const Integrations = () => {
  return (
    <Box
      aria-labelledby="integrations-title"
      as="section"
      className="scroll-mt-24"
      id="integrations"
    >
      <Container className="py-16 md:py-28">
        <Box className="mx-auto max-w-2xl text-center" data-landing-reveal>
          <Text
            as="h2"
            className="text-3xl md:text-5xl"
            id="integrations-title"
          >
            Keep work moving in the tools your team already uses.
          </Text>
          <Text className="text-text-description mx-auto mt-6 max-w-2xl text-base text-pretty">
            Bring requests from Slack, availability from Google and Outlook,
            delivery context from GitHub, and designs from Figma into the same
            plan. Reply to Maya from Gmail.
          </Text>
        </Box>

        <Box
          as="ul"
          className="mx-auto mt-10 flex max-w-[54rem] flex-wrap justify-center gap-3 sm:mt-14 sm:gap-4 md:mt-16"
          data-landing-reveal
          style={{ transitionDelay: "70ms" }}
        >
          {integrations.map((integration) => (
            <Box
              aria-label={`${integration.name} · ${integration.detail}`}
              as="li"
              className="w-[6.5rem] shrink-0 sm:w-28 md:w-32"
              key={integration.name}
            >
              <Box className="bg-surface-muted/80 dark:bg-surface-elevated/85 flex aspect-square items-center justify-center rounded-lg px-4 md:rounded-[1.5rem] md:px-5">
                <Box
                  className="flex size-full items-center justify-center"
                  title={`${integration.name} · ${integration.detail}`}
                >
                  <Image
                    alt=""
                    aria-hidden="true"
                    className={cn(
                      "size-10 object-contain sm:size-11 md:size-14",
                      integration.logoClassName,
                    )}
                    height={56}
                    src={integration.src}
                    width={56}
                  />
                </Box>
              </Box>
            </Box>
          ))}
        </Box>
      </Container>
    </Box>
  );
};
