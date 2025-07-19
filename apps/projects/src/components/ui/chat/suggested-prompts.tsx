import {
  NotificationsIcon,
  StoryIcon,
  SunIcon,
  TeamIcon,
  HelpIcon,
} from "icons";
import { Box, Flex, Wrapper, Text } from "ui";
import { cn } from "lib";
import { usePathname } from "next/navigation";
import { useProfile } from "@/lib/hooks/profile";
import { PriorityIcon } from "../priority-icon";

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
    value: "View all stories and tasks assigned to you across teams.",
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
      <PriorityIcon className="text-danger dark:text-danger" priority="High" />
    ),
    label: "High priority work",
    value: "Find your most urgent stories and tasks to focus on.",
    classes: "bg-danger/10 dark:bg-danger/10",
  },
  {
    icon: <SunIcon className="text-secondary dark:text-white/80" />,
    label: "Switch between light and dark mode",
    value: "Change the app's appearance to match your preference.",
    classes: "bg-secondary/10 dark:bg-secondary/10",
  },
  {
    icon: <HelpIcon className="text-[#6366F1] dark:text-[#6366F1]" />,
    label: "How can you help me?",
    value: "Learn about what I can do and how to use the app effectively.",
    classes: "bg-[#6366F1]/10 dark:bg-[#6366F1]/10",
  },
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  const { data: profile } = useProfile();
  const name = profile?.fullName.split(" ")[0] || profile?.username;
  const pathname = usePathname();
  const isMaya = pathname.includes("maya");
  return (
    <Box
      className={cn("px-12 py-4", {
        "md:px-10 md:py-6": isMaya,
      })}
    >
      <Text
        className={cn("text-center text-xl font-semibold", {
          "mx-auto mb-10 md:w-11/12 md:text-5xl": isMaya,
        })}
      >
        Hi, {name}! How can Maya help you today?
      </Text>
      <Text
        className={cn("mx-auto mt-3 text-center", {
          "md:w-11/12 md:text-lg": isMaya,
        })}
        color="muted"
      >
        I&apos;m here to help you manage your work, stay organized, and keep
        your projects moving. Choose a suggestion below or ask me anything!
      </Text>
      <Box
        className={cn("mt-5 flex flex-col gap-3", {
          "grid grid-cols-2 gap-4": isMaya,
        })}
      >
        {SUGGESTED_PROMPTS.map((prompt, index) => (
          <Wrapper
            className={cn(
              "flex cursor-pointer items-center gap-2 py-3.5 ring-primary transition hover:ring-2 md:px-3",
              {
                "gap-3 md:px-4 md:py-4": isMaya,
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
                "size-10 shrink-0 rounded-lg bg-gray-50 dark:bg-dark-200",
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
                  "md:text-lg": isMaya,
                })}
              >
                {prompt.label}
              </Text>
              <Text
                className={cn("text-[0.95rem]", {
                  "md:text-base": isMaya,
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
