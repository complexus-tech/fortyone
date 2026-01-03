import { NotificationsIcon, StoryIcon, TeamIcon } from "icons";
import { Box, Flex, Wrapper, Text } from "ui";
import { cn } from "lib";
import { useProfile } from "@/lib/hooks/profile";
import { PriorityIcon } from "../priority-icon";

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
  isOnPage?: boolean;
  fromIndex?: number;
};

export const SuggestedPrompts = ({
  onPromptSelect,
  isOnPage,
  fromIndex = 0,
}: SuggestedPromptsProps) => {
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
      value: "View all stories assigned to you across teams.",
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
      value: "Find your most urgent stories across teams to focus on.",
      classes: "bg-danger/10 dark:bg-danger/10",
    },
  ];
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
        {SUGGESTED_PROMPTS.slice(fromIndex).map((prompt, index) => (
          <Wrapper
            className={cn(
              "flex cursor-pointer items-center gap-3 ring-primary transition hover:ring-2 md:px-4",
              {
                "gap-3 md:px-4 md:py-3": isOnPage,
              },
            )}
            key={index}
            onClick={() => {
              onPromptSelect(prompt.label);
            }}
            tabIndex={0}
          >
            <Flex
              align="center"
              className={cn(
                "size-10 shrink-0 rounded-lg bg-surface-muted",
                prompt.classes,
              )}
              gap={2}
              justify="center"
            >
              {prompt.icon}
            </Flex>
            <Box>
              <Text
                className={cn({
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
