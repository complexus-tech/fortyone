import React from "react";
import { Box, Button, Text, Wrapper } from "ui";

export const LimitReached = ({ isOnPage }: { isOnPage?: boolean }) => {
  return (
    <Box className="mb-4 px-6">
      <Wrapper className="flex items-center justify-between gap-4">
        <Text>
          You've reached your monthly AI chat limit. Messages reset on the 1st,
          or upgrade to continue chatting with Maya! ✨
        </Text>
        <Button
          color="invert"
          className="shrink-0"
          href="/settings/workspace/billing"
        >
          Upgrade {isOnPage && "plan"}
        </Button>
      </Wrapper>
    </Box>
  );
};
