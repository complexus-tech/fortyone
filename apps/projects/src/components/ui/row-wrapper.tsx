import { cn } from "lib";
import type { ReactNode } from "react";
import { Container } from "ui";

export const RowWrapper = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => {
  return (
    <Container
      className={cn(
        "group flex items-center justify-between border-b-[0.5px] border-gray-50 py-2.5 outline-none hover:bg-gray-50/30 focus:bg-gray-50/50 dark:border-dark-100/70 hover:dark:bg-dark-200/30 focus:dark:bg-dark-200/50",
        className,
      )}
      tabIndex={0}
    >
      {children}
    </Container>
  );
};
