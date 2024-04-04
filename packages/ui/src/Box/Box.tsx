import { CSSProperties, HTMLAttributes, ReactNode, createElement } from "react";

export interface BoxProps extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  as?: keyof JSX.IntrinsicElements;
  style?: CSSProperties;
  html?: string;
  children?: ReactNode;
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
