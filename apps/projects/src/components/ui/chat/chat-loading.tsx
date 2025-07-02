import { Flex } from "ui";
import { AiIcon } from "./ai";
import { Thinking } from "./thinking";

export const ChatLoading = () => {
  return (
    <Flex gap={3}>
      <AiIcon />
      <Thinking />
    </Flex>
  );
};
