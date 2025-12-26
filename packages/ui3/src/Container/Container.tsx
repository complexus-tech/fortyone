import { cn } from "lib";
import { JSX } from "react";
import { CSSProperties, HTMLAttributes, ReactNode, createElement } from "react";

export interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  id?: string;
  style?: CSSProperties;
  children?: ReactNode;
  as?: keyof JSX.IntrinsicElements;
  full?: boolean;
}

export const Container = ({
  className = "",
  style = {},
  as = "div",
  id,
  children,
  full,
  ...rest
}: ContainerProps) => {
  return createElement(as, {
    className: cn(
      "mx-auto px-5 md:px-12",
      {
        "mx-auto px-0": full,
      },
      className
    ),
    id,
    style,
    ...rest,
    children,
  });
};
