import { NotificationsIcon, StoryIcon, SunIcon, TeamIcon } from "icons";
import { Box, Flex, Wrapper, Text } from "ui";
import { cn } from "lib";
import { useProfile } from "@/lib/hooks/profile";
import { PriorityIcon } from "../priority-icon";

const SUGGESTED_PROMPTS = [
  {
    icon: <TeamIcon className="text-success dark:text-success" />,
    label: "Who’s on my team?",
    value: "See a list of everyone in your current team.",
    classes: "bg-success/10 dark:bg-success/10",
  },
  {
    icon: <StoryIcon className="text-warning dark:text-warning" />,
    label: "What’s on my plate?",
    value: "View all stories and tasks assigned to you.",
    classes: "bg-warning/10 dark:bg-warning/10",
  },
  {
    icon: <NotificationsIcon className="text-info dark:text-info" />,
    label: "What’s new for me?",
    value: "Check your latest unread notifications.",
    classes: "bg-info/10 dark:bg-info/10",
  },
  {
    icon: (
      <PriorityIcon className="text-danger dark:text-danger" priority="High" />
    ),
    label: "High priority work",
    value: "Find your most urgent stories and tasks.",
    classes: "bg-danger/10 dark:bg-danger/10",
  },
  {
    icon: <SunIcon className="text-secondary dark:text-white/80" />,
    label: "Switch between light and dark mode",
    value: "Change the app’s appearance to match your preference.",
    classes: "bg-secondary/10 dark:bg-secondary/10",
  },
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  const { data: profile } = useProfile();
  const name = profile?.fullName.split(" ")[0] || profile?.username;
  return (
    <Box className="px-6 py-4">
      <Text className="text-center text-xl font-semibold">
        Hi, {name}! How can Maya help you today?
      </Text>
      <Text className="mx-auto mt-3 text-center" color="muted">
        I&apos;m here to help you manage your work, stay organized, and keep
        your projects moving. Choose a suggestion below or ask me anything!
      </Text>
      <Flex className="mt-5" direction="column" gap={3}>
        {SUGGESTED_PROMPTS.map((prompt, index) => (
          <Wrapper
            className="flex cursor-pointer items-center gap-2 py-3.5 ring-primary transition hover:ring-2 md:px-3"
            key={index}
            onClick={() => {
              onPromptSelect(prompt.label);
            }}
            tabIndex={0}
          >
            <Flex
              align="center"
              className={cn(
                "size-10 rounded-lg bg-gray-50 dark:bg-dark-200",
                prompt.classes,
              )}
              gap={2}
              justify="center"
            >
              {prompt.icon}
            </Flex>
            <Box>
              <Text>{prompt.label}</Text>
              <Text className="text-[0.95rem]" color="muted">
                {prompt.value}
              </Text>
            </Box>
          </Wrapper>
        ))}
      </Flex>
    </Box>
  );
};
