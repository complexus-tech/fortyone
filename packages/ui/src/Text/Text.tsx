import { VariantProps, cva } from 'cva';
import { FC, HTMLAttributes, JSXElementConstructor } from 'react';

import { cn } from 'lib';

const text = cva('text-black dark:text-white', {
  variants: {
    align: {
      left: 'text-left',
      center: 'text-center',
      right: 'text-right',
    },
    color: {
      primary: 'text-primary-200',
      muted: 'text-gray-300/60 dark:text-gray-200/80',
      danger: 'text-danger',
      black: 'text-black',
      white: 'text-white',
      warning: 'text-warning',
      info: 'text-info',
      secondary: 'text-secondary',
    },
    fontSize: {
      xs: 'text-xs',
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-lg',
      xl: 'text-xl',
      '2xl': 'text-2xl',
      '3xl': 'text-3xl',
      '4xl': 'text-4xl',
      inherit: 'text-inherit',
    },
    fontWeight: {
      light: 'font-light',
      normal: 'font-normal',
      medium: 'font-medium',
      semibold: 'font-semibold',
      bold: 'font-bold',
    },
    transform: {
      uppercase: 'uppercase',
      lowercase: 'lowercase',
      capitalize: 'capitalize',
      none: 'normal-case',
    },
    fontStyle: {
      italic: 'italic',
      normal: 'not-italic',
    },
    decoration: {
      underline: 'underline',
      lineThrough: 'line-through',
      none: 'no-underline',
    },
    textOverflow: {
      ellipsis: 'text-ellipsis',
      clip: 'text-clip',
      truncate: 'truncate',
    },
  },
});

export interface TextProps
  extends Omit<
      HTMLAttributes<
        HTMLHeadingElement | HTMLSpanElement | HTMLParagraphElement
      >,
      'color'
    >,
    VariantProps<typeof text> {
  as?: keyof JSX.IntrinsicElements | JSXElementConstructor<any>;
  html?: string;
}

export const Text: FC<TextProps> = ({
  as: Tag = 'p',
  children,
  className,
  html,
  align,
  color,
  fontSize,
  fontWeight,
  transform,
  fontStyle,
  decoration,
  textOverflow,
  ...rest
}) => {
  const htmlProps = html
    ? {
        dangerouslySetInnerHTML: { __html: html },
      }
    : {};

  const classes = cn(
    text({
      align,
      color,
      fontSize,
      fontWeight,
      transform,
      fontStyle,
      decoration,
      textOverflow,
    }),
    className
  );
  return (
    <Tag className={classes} {...rest} {...htmlProps}>
      {children}
    </Tag>
  );
};
