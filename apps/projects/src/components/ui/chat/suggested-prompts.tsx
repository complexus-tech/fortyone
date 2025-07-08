import { Button, Box, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Toggle color theme",
  "List all team members",
  "Show me my assigned stories",
  "Show me unread notifications",
  "Show me high priority items",
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  return (
    <Box className="p-6">
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
