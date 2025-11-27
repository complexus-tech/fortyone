import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick top-0 z-10 flex h-[3.6rem] w-full items-center border-b-[0.5px] border-gray-200/60 dark:border-dark-100",
        className,
      )}
    >
      {children}
    </Container>
  );
};
