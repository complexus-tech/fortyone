import type { ReactNode } from "react";
import { Flex } from "ui";

export const KanbanList = ({ children }: { children: ReactNode }) => (
  <Flex
    className="h-[calc(100vh-7.5rem)] overflow-y-auto pb-6"
    direction="column"
    gap={3}
  >
    {children}
  </Flex>
);
