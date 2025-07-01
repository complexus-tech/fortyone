import { Button, Box, Text, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Show me my assigned stories",
  "Show me high priority items",
  "Show me unread notifications",
  "List all team members",
  "Toggle color theme",
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  return (
    <Box className="px-6 pt-6">
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
            rounded="lg"
          >
            {prompt}
          </Button>
        ))}
      </Flex>
    </Box>
  );
};
