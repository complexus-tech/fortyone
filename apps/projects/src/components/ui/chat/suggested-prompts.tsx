import { Button, Box, Text, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Show me my assigned stories",
  "Take me to settings",
  "Create a new story",
  "Change system theme",
  "Create a new objective",
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
            rounded="full"
          >
            {prompt}
          </Button>
        ))}
      </Flex>
    </Box>
  );
};
