import { Button, Box, Text, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Show me my assigned stories",
  "Get current sprint summary",
  "Navigate to analytics",
  "Create a new story",
  "What's the team velocity?",
  "List active sprints",
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
