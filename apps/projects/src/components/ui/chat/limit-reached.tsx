import React from "react";
import { Box, Button, Text, Wrapper } from "ui";

export const LimitReached = () => {
  return (
    <Box className="mb-4 px-6">
      <Wrapper className="flex items-center gap-4">
        <Text>
          Lorem ipsum dolor sit amet consectetur adipisicing elit. Voluptates
          sint corrupti quae.
        </Text>
        <Button color="invert">Upgrade</Button>
      </Wrapper>
    </Box>
  );
};
