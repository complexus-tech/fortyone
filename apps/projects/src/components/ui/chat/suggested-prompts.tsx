import { Notification02Icon, StoryIcon, TeamIcon } from "icons";
import { Box, Flex, Text } from "ui";
import { cn } from "lib";
import { useProfile } from "@/lib/hooks/profile";
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
  const { data: profile } = useProfile();
  const name = profile?.fullName.split(" ")[0] || profile?.username;

  const SUGGESTED_PROMPTS = [
    {
      icon: (
        <TeamIcon className="text-success dark:text-success size-[1.1875rem]" />
      ),
      label: "Plan project",
      value: "Build a clear project plan.",
    },
    {
      icon: (
        <StoryIcon className="text-warning dark:text-warning size-[1.1875rem]" />
      ),
      label: "Sprint summary",
      value: "Summarize the current sprint.",
    },
    {
      icon: (
        <Notification02Icon className="text-info dark:text-info size-[1.1875rem]" />
      ),
      label: "Status report",
      value: "Draft a concise project update.",
    },
    {
      icon: (
        <PriorityIcon
          className="text-danger dark:text-danger size-[1.1875rem]"
          priority="High"
        />
      ),
      label: "Active work",
      value: "Find all work currently in progress.",
    },
  ];

  if (isPopup) {
    return (
      <Box className="px-[18px] pt-8 pb-[18px]">
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
    <Box className={cn("px-6 py-4", { "px-0 pt-7 pb-0": isOnPage })}>
      <Box
        className={cn("mt-6 flex flex-col gap-3", {
          "mt-0 grid sm:grid-cols-2 sm:gap-3 lg:grid-cols-4": isOnPage,
        })}
      >
        {SUGGESTED_PROMPTS.slice(fromIndex).map((prompt) => (
          <button
            className={cn(
              "border-border/70 bg-surface/72 hover:bg-surface focus-visible:ring-primary dark:bg-surface/55 dark:hover:bg-surface/80 flex min-h-[6.25rem] cursor-pointer flex-col items-start gap-2.5 rounded-2xl border px-3.5 py-3 text-left transition-[background-color,border-color,transform] hover:-translate-y-0.5 hover:border-[color-mix(in_oklab,var(--color-info)_35%,var(--color-border))] focus-visible:ring-2 focus-visible:outline-none",
              {
                "sm:min-h-[6.5rem] lg:px-4 lg:py-3": isOnPage,
              },
            )}
            key={prompt.label}
            onClick={() => {
              onPromptSelect(prompt.label);
            }}
            type="button"
          >
            <Flex align="center" className="size-5 shrink-0" justify="center">
              {prompt.icon}
            </Flex>
            <Box className="w-full min-w-0 overflow-hidden">
              <Text
                className={cn("font-semibold", {
                  "text-base leading-5": isOnPage,
                })}
              >
                {prompt.label}
              </Text>
              <Text
                className={cn("text-[0.95rem]", {
                  "mt-1 block w-full overflow-hidden leading-5 text-ellipsis whitespace-nowrap":
                    isOnPage,
                })}
                color="muted"
              >
                {prompt.value}
              </Text>
            </Box>
          </button>
        ))}
      </Box>
    </Box>
  );
};
