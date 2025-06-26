import { Button, Box, Text, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Summarize the current sprint progress",
  "What are the highest priority tasks?",
  "Show me blocked stories and their blockers",
  "Generate a status report for the team",
  "What's the team velocity this sprint?",
  "Help me plan the next sprint",
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  return (
    <Box className="border-t border-gray-100 px-6 py-6 dark:border-dark-100">
      <Text className="mb-4" fontSize="md" fontWeight="medium">
        Try asking:
      </Text>
      <Flex gap={3} wrap>
        {SUGGESTED_PROMPTS.map((prompt, index) => (
          <Button
            color="tertiary"
            key={index}
            onClick={() => {
              onPromptSelect(prompt);
            }}
            rounded="full"
            variant="outline"
          >
            {prompt}
          </Button>
        ))}
      </Flex>
    </Box>
  );
};
