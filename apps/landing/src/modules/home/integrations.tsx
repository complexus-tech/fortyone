import Image from "next/image";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";

type Integration = {
  logoClassName?: string;
  name: string;
  src: string;
};

const integrations: readonly Integration[] = [
  { name: "Slack", src: "/integrations/slack.svg" },
  { name: "Intercom", src: "/integrations/intercom-icon.svg" },
  {
    name: "Notion",
    src: "/integrations/notion.svg",
    logoClassName: "invert dark:invert-0",
  },
  { name: "Figma", src: "/integrations/figma.svg" },
  {
    name: "GitHub",
    src: "/integrations/github.svg",
    logoClassName: "dark:invert",
  },
  { name: "Jira", src: "/integrations/jira.svg" },
  { name: "Google Drive", src: "/integrations/drive.svg" },
  { name: "GitLab", src: "/integrations/gitlab.svg" },
  { name: "Microsoft Teams", src: "/integrations/teams.svg" },
  {
    name: "Zendesk",
    src: "/integrations/zend.svg",
    logoClassName: "invert dark:invert-0",
  },
];

export const Integrations = () => {
  return (
    <Box className="scroll-mt-24" id="integrations">
      <Container className="py-16 md:py-28">
        <Box className="grid gap-12 lg:grid-cols-[minmax(0,0.8fr)_minmax(0,1.2fr)] lg:items-start lg:gap-20 xl:gap-28">
          <Box data-landing-reveal>
            <Text
              as="h2"
              className="max-w-xl pb-1 text-4xl text-balance md:text-5xl"
            >
              Keep project work connected across the tools your team uses.
            </Text>
            <Text className="mt-6 max-w-lg leading-relaxed text-pretty opacity-70">
              Use Google Calendar for availability and Slack for task intake.
              GitHub, Figma, and GitLab bring delivery context into the project
              plan.
            </Text>
          </Box>

          <Box data-landing-reveal style={{ transitionDelay: "70ms" }}>
            <Box
              as="ul"
              className="border-border/80 grid grid-cols-2 border-r-[0.5px] border-b-[0.5px] sm:grid-cols-5"
            >
              {integrations.map((integration) => (
                <Box
                  as="li"
                  className="border-border/80 bg-background/40 flex min-h-28 flex-col items-center justify-center gap-3 border-t-[0.5px] border-l-[0.5px] px-4 py-6"
                  key={integration.name}
                >
                  <Image
                    alt=""
                    aria-hidden="true"
                    className={cn(
                      "size-9 object-contain opacity-80 grayscale md:size-10",
                      integration.logoClassName,
                    )}
                    height={40}
                    src={integration.src}
                    width={40}
                  />
                  <Text className="text-center text-sm leading-tight opacity-60">
                    {integration.name}
                  </Text>
                </Box>
              ))}
            </Box>
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
