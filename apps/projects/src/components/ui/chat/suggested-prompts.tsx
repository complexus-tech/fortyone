import { NotificationsIcon, StoryIcon, TeamIcon } from "icons";
import { Box, Flex, Wrapper, Text } from "ui";
import { cn } from "lib";
import { useProfile } from "@/lib/hooks/profile";
import { useTerminology } from "@/hooks";
import { PriorityIcon } from "../priority-icon";

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
  isOnPage?: boolean;
  isPopup?: boolean;
  fromIndex?: number;
};

const POPUP_PROMPTS = [
  "What should I focus on today?",
  "What changed since I last checked?",
  "Which work is at risk?",
  "Show me the current sprint",
  "What is blocking my team?",
  "Help me plan my highest-priority task",
];

export const SuggestedPrompts = ({
  onPromptSelect,
  isOnPage,
  isPopup = false,
  fromIndex = 0,
}: SuggestedPromptsProps) => {
  const { getTermDisplay } = useTerminology();
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const { data: profile } = useProfile();
  const name = profile?.fullName.split(" ")[0] || profile?.username;

  const SUGGESTED_PROMPTS = [
    {
      icon: <TeamIcon className="text-success dark:text-success" />,
      label: "Who's on my team?",
      value: "See a list of everyone in your current team and their roles.",
      classes: "bg-success/10 dark:bg-success/10",
    },
    {
      icon: <StoryIcon className="text-warning dark:text-warning" />,
      label: "What's on my plate?",
      value: `View all ${storyTermPlural} assigned to you across teams.`,
      classes: "bg-warning/10 dark:bg-warning/10",
    },
    {
      icon: <NotificationsIcon className="text-info dark:text-info" />,
      label: "What's new for me?",
      value: "Check your latest unread notifications and updates.",
      classes: "bg-info/10 dark:bg-info/10",
    },
    {
      icon: (
        <PriorityIcon
          className="text-danger dark:text-danger"
          priority="High"
        />
      ),
      label: "High priority work",
      value: `Find your most urgent ${storyTermPlural} across teams to focus on.`,
      classes: "bg-danger/10 dark:bg-danger/10",
    },
  ];

  if (isPopup) {
    return (
      <Box className="px-[18px] pt-14 pb-[18px]">
        <Text as="h2" className="text-4xl leading-[1.08] tracking-[-0.04em]">
          Hi, {name}! Ask me anything!
        </Text>
        <Text className="mt-3 max-w-sm text-base leading-6" color="muted">
          Ask about priorities, delivery risks, your team, or what to do next.
        </Text>
        <Box className="mt-[22px] border-t border-black/[0.07] dark:border-white/[0.07]">
          {POPUP_PROMPTS.slice(fromIndex).map((prompt) => (
            <button
              className="text-foreground hover:text-primary focus-visible:ring-primary flex min-h-[52px] w-full items-center border-0 border-b border-black/[0.07] bg-transparent px-px py-3 text-left text-[1.1rem] leading-[1.4rem] transition-colors focus-visible:ring-2 focus-visible:outline-none dark:border-white/[0.07]"
              key={prompt}
              onClick={() => {
                onPromptSelect(prompt);
              }}
              type="button"
            >
              {prompt}
            </button>
          ))}
        </Box>
      </Box>
    );
  }

  return (
    <Box
      className={cn("px-6 py-4", {
        "md:py-6": isOnPage,
      })}
    >
      <Text
        className={cn("mx-auto w-max pb-1.5 text-center text-4xl md:w-11/12", {
          "md:mb-10 md:text-5xl": isOnPage,
        })}
      >
        Hi, {name}! Ask me anything!
      </Text>
      <Box
        className={cn("mt-6 flex flex-col gap-3", {
          "grid md:grid-cols-2 md:gap-4": isOnPage,
        })}
      >
        {SUGGESTED_PROMPTS.slice(fromIndex).map((prompt) => (
          <Wrapper
            className={cn(
              "dark:bg-surface/60 ring-primary flex cursor-pointer items-center gap-3 transition hover:ring-2 md:px-4",
              {
                "gap-3 md:px-4 md:py-3": isOnPage,
              },
            )}
            key={prompt.label}
            onClick={() => {
              onPromptSelect(prompt.label);
            }}
            tabIndex={0}
          >
            <Flex
              align="center"
              className={cn(
                "bg-surface-muted size-10 shrink-0 rounded-lg",
                prompt.classes,
              )}
              gap={2}
              justify="center"
            >
              {prompt.icon}
            </Flex>
            <Box>
              <Text
                className={cn("font-semibold", {
                  "md:text-lg": isOnPage,
                })}
              >
                {prompt.label}
              </Text>
              <Text
                className={cn("text-[0.95rem]", {
                  "md:mt-0.5 md:text-base md:leading-[1.3rem]": isOnPage,
                })}
                color="muted"
              >
                {prompt.value}
              </Text>
            </Box>
          </Wrapper>
        ))}
      </Box>
    </Box>
  );
};
