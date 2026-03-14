import { cn } from "lib";
import type { ContainerProps } from "ui";
import { Container as UContainer } from "ui";

export const Container = ({ className, ...rest }: ContainerProps) => {
  return <UContainer className={cn("max-w-350", className)} {...rest} />;
};
