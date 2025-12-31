import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick border-border top-0 z-10 flex h-[3.6rem] w-full items-center border-b-[0.5px]",
        className,
      )}
    >
      {children}
    </Container>
  );
};
