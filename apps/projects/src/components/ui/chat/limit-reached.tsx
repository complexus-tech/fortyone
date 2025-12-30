import React from "react";
import { Box, Button, Text, Wrapper } from "ui";

export const LimitReached = () => {
  return (
    <Box className="mb-4 px-6">
      <Wrapper className="flex items-center gap-4">
        <Text>
          Message limit hit! Resets on the 1st, or upgrade to keep the Maya
          magic flowing ✨
        </Text>
        <Button
          color="invert"
          className="shrink-0"
          href="/settings/workspace/billing"
        >
          Upgrade
        </Button>
      </Wrapper>
    </Box>
  );
};
