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
  as: Tag = "div",
  children,
  html,
  ...rest
}: BoxProps) => {
  const htmlProps = html
    ? {
        dangerouslySetInnerHTML: { __html: html },
      }
    : {};

  return createElement(Tag, {
    className,
    style,
    ...rest,
    ...htmlProps,
    children,
  });
};
