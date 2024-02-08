import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick top-0 z-10 flex h-16 w-full items-center border-b border-gray-100 bg-white/50 backdrop-blur dark:border-dark-200 dark:bg-dark/70",
        className,
      )}
    >
      {children}
    </Container>
  );
};
