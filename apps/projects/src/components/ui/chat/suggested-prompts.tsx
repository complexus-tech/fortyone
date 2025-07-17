import { AiIcon, NotificationsIcon, StoryIcon, SunIcon, TeamIcon } from "icons";
import { Box, Flex, Wrapper, Text } from "ui";
import { useProfile } from "@/lib/hooks/profile";
import { PriorityIcon } from "../priority-icon";

const SUGGESTED_PROMPTS = [
  {
    icon: <SunIcon />,
    label: "Toggle color theme",
    value: "Update the system theme",
  },
  {
    icon: <TeamIcon />,
    label: "List all team members",
    value: "Show me all team members in your team",
  },
  {
    icon: <StoryIcon />,
    label: "Show me my assigned stories",
    value: "Show me my assigned stories",
  },
  {
    icon: <NotificationsIcon />,
    label: "Show me unread notifications",
    value: "Show me unread notifications",
  },
  {
    icon: <PriorityIcon priority="High" />,
    label: "Show me high priority items",
    value: "Show me high priority items",
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
      <Flex align="center" className="mx-auto" gap={2} justify="center">
        <AiIcon className="h-11 text-primary" />
      </Flex>
      <Text className="mt-5 text-center text-xl font-semibold">
        Hi, {name}! Plan your day with Maya
      </Text>
      <Text className="mt-3 text-center" color="muted">
        I&apos;m here to help you plan your day. Lorem ipsum dolor sit amet
        consectetur adipisicing elit. Veniam, ullam!
      </Text>
      <Flex className="mt-10" direction="column" gap={3}>
        {SUGGESTED_PROMPTS.map((prompt, index) => (
          <Wrapper
            className="flex cursor-pointer items-center gap-2 ring-primary transition hover:ring-2 md:px-4"
            key={index}
            onClick={() => {
              onPromptSelect(prompt.label);
            }}
            tabIndex={0}
          >
            <Flex
              align="center"
              className="size-10 rounded-lg bg-gray-50 dark:bg-dark-200"
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
