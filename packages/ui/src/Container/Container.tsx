import { cn } from 'lib';
import {
  CSSProperties,
  FC,
  HTMLAttributes,
  JSXElementConstructor,
  ReactNode,
} from 'react';

export interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  className?: string;
  id?: string;
  style?: CSSProperties;
  children?: ReactNode;
  as?:
    | 'div'
    | 'section'
    | 'nav'
    | 'header'
    | 'footer'
    | JSXElementConstructor<any>;
  full?: boolean;
}

export const Container: FC<ContainerProps> = ({
  className = '',
  style = {},
  as: Tag = 'div',
  id,
  children,
  full,
  ...rest
}) => {
  return (
    <Tag
      className={cn(
        'mx-auto px-10',
        {
          'mx-auto px-0': full,
        },
        className
      )}
      id={id}
      style={style}
      {...rest}
    >
      {children}
    </Tag>
  );
};
