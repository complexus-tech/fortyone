import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container } from "ui";

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        "stick border-border top-0 z-10 flex h-(--app-page-header-height) w-full shrink-0 items-center border-b-[0.5px]",
        className,
      )}
      data-header-container
    >
      {children}
    </Container>
  );
};
