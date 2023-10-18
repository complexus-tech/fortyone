import type { ReactNode } from "react";
import { Container } from "ui";

export const RowWrapper = ({ children }: { children: ReactNode }) => {
  return (
    <Container
      className="group flex
       items-center justify-between border-b border-gray-50 py-3 outline-none hover:bg-gray-50/50 focus:bg-gray-50/50 dark:border-dark-200 hover:dark:bg-dark-200/50 focus:dark:bg-dark-200/50"
      tabIndex={0}
    >
      {children}
    </Container>
  );
};
