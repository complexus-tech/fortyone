import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "fixed right-0 top-0 z-10 flex h-16 w-[calc(100vw-220px)] items-center border-b border-gray-100 bg-white/50 backdrop-blur dark:border-dark-100 dark:bg-dark/50",
        className,
      )}
    >
      {children}
    </Container>
  );
};
