import type { VariantProps } from 'cva';
import { cva } from 'cva';
import { cn } from 'lib';
import Link from 'next/link';
import {
  forwardRef,
  type ButtonHTMLAttributes,
  type JSXElementConstructor,
  type ReactElement,
} from 'react';
import { AiOutlineLoading3Quarters } from 'react-icons/ai';

const button = cva(
  'flex w-max items-center border gap-2 transition duration-200 ease-linear focus:outline-0',
  {
    variants: {
      variant: {
        outline: null,
        solid: null,
        naked: 'bg-opacity-10',
      },
      rounded: {
        none: 'rounded-none',
        sm: 'rounded',
        md: 'rounded-lg',
        lg: 'rounded-xl',
        xl: 'rounded-3xl',
        full: 'rounded-full',
      },
      color: {
        primary:
          'border-primary bg-primary ring-primary-200 enabled:hover:bg-primary-200 enabled:hover:border-primary-200 focus:bg-primary-200 focus:border-primary-200',
        danger:
          'text-danger border-danger bg-danger ring-danger enabled:hover:bg-danger-300 enabled:hover:border-danger-300 focus:bg-danger-300 focus:border-danger-300',
        info: 'text-info border-info bg-info ring-info enabled:hover:bg-info-300 enabled:hover:border-info-300 focus:bg-info-300 focus:border-info-300',
        warning:
          'text-warning border-warning bg-warning ring-warning enabled:hover:bg-warning-300 enabled:hover:border-warning-300 focus:bg-warning-300 focus:border-warning-300',
        tertiary:
          'text-gray-300 dark:text-gray-200 dark:border-dark-100 dark:bg-dark-200/50 bg-gray-100 border-gray-100 focus:bg-gray-50 dark:focus:bg-dark-200 hover:bg-gray-50 dark:hover:bg-dark-200',
        secondary:
          'text-secondary border-secondary bg-secondary ring-secondary',
      },
      size: {
        xs: 'pr-2 pl-1 h-[1.85rem] text-[0.95rem] gap-[2px]',
        sm: 'px-2 h-[2.2rem] gap-1',
        md: 'px-3 h-[2.5rem]',
        lg: 'px-7 py-3 font-semibold',
      },
      disabled: {
        true: 'opacity-40 cursor-not-allowed',
      },
      active: {
        true: null,
      },
      loading: {
        true: 'opacity-80 cursor-progress',
      },
      fullWidth: {
        true: 'w-full',
      },
      align: {
        center: 'justify-center',
        left: 'justify-start',
        right: 'justify-end',
        between: 'justify-between',
      },
    },
    compoundVariants: [
      // Solid variant
      {
        variant: 'solid',
        color: ['primary', 'secondary', 'warning', 'danger', 'info'],
        className: 'text-white',
      },
      // Outline variant
      {
        variant: 'outline',
        color: ['primary', 'secondary', 'warning', 'danger', 'info'],
        className:
          'enabled:bg-opacity-0 hover:enabled:bg-opacity-10 focus:enabled:bg-opacity-10',
      },
      {
        variant: 'outline',
        color: 'tertiary',
        className: 'bg-gray-50/10',
      },
      // Naked variant
      {
        variant: 'naked',
        color: [
          'primary',
          'secondary',
          'warning',
          'danger',
          'info',
          'tertiary',
        ],
        className: 'bg-transparent dark:bg-transparent border-none',
      },
      {
        variant: 'naked',
        color: ['secondary'],
        className: 'dark:text-white',
      },
      {
        disabled: true,
        variant: ['outline', 'naked'],
        className: 'bg-opacity-100 text-white',
      },
      {
        disabled: true,
        variant: ['outline', 'naked'],
        color: ['primary', 'tertiary'],
        className: 'bg-opacity-100 text-black',
      },
    ],
    defaultVariants: {
      size: 'md',
      variant: 'solid',
      color: 'primary',
      rounded: 'md',
      align: 'left',
    },
  }
);

interface ButtonProps
  extends Omit<
      ButtonHTMLAttributes<HTMLButtonElement>,
      'color' | 'disabled' | 'active'
    >,
    VariantProps<typeof button> {
  href?: string;
  loadingText?: string;
  target?: '_blank' | '_self' | '_parent' | '_top';
  as?: 'button' | 'a' | JSXElementConstructor<any>;
  rightIcon?: ReactElement;
  leftIcon?: ReactElement;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (props, ref) => {
    const {
      as: Tag = 'button',
      variant,
      color,
      rounded,
      size,
      target = '_self',
      loading,
      fullWidth,
      align,
      loadingText,
      href,
      active,
      rightIcon,
      leftIcon,
      className,
      disabled,
      children,
      ...rest
    } = props;

    const classes = cn(
      button({
        variant,
        color,
        size,
        disabled,
        loading,
        active,
        rounded,
        fullWidth,
        align,
      }),
      className
    );

    return (
      <>
        {href ? (
          <Link className={classes} href={href} target={target}>
            {leftIcon}
            {children}
            {rightIcon}
          </Link>
        ) : (
          <Tag className={classes} disabled={disabled} ref={ref} {...rest}>
            {loading ? (
              <>
                <AiOutlineLoading3Quarters className='animate-spin' />
                {loadingText || 'Loading...'}
              </>
            ) : (
              <>
                {leftIcon}
                {children}
                {rightIcon}
              </>
            )}
          </Tag>
        )}
      </>
    );
  }
);

Button.displayName = 'Button';
