import { Button, Box, Flex } from "ui";

const SUGGESTED_PROMPTS = [
  "Show me high priority items",
  "Toggle color theme",
  "List all team members",
  "Show me my assigned stories",
  "Show me unread notifications",
];

type SuggestedPromptsProps = {
  onPromptSelect: (prompt: string) => void;
};

export const SuggestedPrompts = ({ onPromptSelect }: SuggestedPromptsProps) => {
  return (
    <Box className="p-6">
      <Flex gap={3} justify="center" wrap>
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
