import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick top-0 z-10 flex h-16 w-full items-center border-b-[0.5px] border-gray-100 dark:border-dark-50",
        className,
      )}
    >
      {children}
    </Container>
  );
};
