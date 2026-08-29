import { createElement } from "react";
import type { CSSProperties, HTMLAttributes, JSX, ReactNode, Ref } from "react";

export interface BoxProps extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  as?: keyof JSX.IntrinsicElements;
  style?: CSSProperties;
  html?: string;
  children?: ReactNode;
  ref?: Ref<HTMLDivElement>;
}

export const Box = ({
  className = "",
  style = {},
  as = "div",
  children,
  html,
  ...rest
}: BoxProps) => {
  const htmlProps = html
    ? {
        dangerouslySetInnerHTML: { __html: html },
      }
    : {};

  return createElement(as, {
    className,
    style,
    ...rest,
    ...htmlProps,
    children,
  });
};
