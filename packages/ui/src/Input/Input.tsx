import {
  forwardRef,
  ReactElement,
  InputHTMLAttributes,
  ReactNode,
} from "react";
import { cva, type VariantProps } from "cva";
import { cn } from "lib";

const inputVariants = cva(
  "w-full rounded-[0.45rem] border bg-white/70 dark:bg-dark/20 border-gray-100 dark:border-dark-100 dark:ring-offset-dark px-4 h-[2.8rem] leading-[2.8rem] focus:outline-0 focus:ring-[2.5px] focus:ring-gray-100 dark:focus:ring-dark-50 focus:ring-offset-1 read-only:focus:ring-0 placeholder:text-gray/80 dark:placeholder:text-gray-300",
  {
    variants: {
      size: {
        sm: "h-[2rem] text-sm px-3",
        md: "h-[2.8rem]", // default size from original component
        lg: "h-[3.2rem] focus:ring-[3px]",
      },
      rounded: {
        none: "rounded-none",
        sm: "rounded-md",
        md: "rounded-[0.45rem]", // default from original
        lg: "rounded-xl",
        xl: "rounded-2xl",
        full: "rounded-full",
      },
      variant: {
        default: "", // uses the base styles
        solid: "bg-gray-200 border-0 dark:bg-dark-300 dark:border-dark-200",
        error:
          "border-danger dark:border-danger focus:ring-danger dark:focus:ring-danger",
      },
    },
    defaultVariants: {
      size: "md",
      variant: "default",
      rounded: "md",
    },
  }
);

export interface InputProps
  extends Omit<InputHTMLAttributes<HTMLInputElement>, "size">,
    VariantProps<typeof inputVariants> {
  label?: string;
  helpText?: string;
  hasError?: boolean;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  labelClassName?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>((props, ref) => {
  const {
    label,
    className,
    required,
    value,
    helpText,
    type = "text",
    hasError,
    size,
    variant,
    rounded,
    leftIcon,
    rightIcon,
    labelClassName,
    ...rest
  } = props;

  const inputClasses = cn(
    inputVariants({
      size,
      variant: hasError ? "error" : variant,
      rounded,
    }),
    className
  );

  return (
    <label className="relative block">
      {label && (
        <span className={cn("mb-[0.35rem] block", labelClassName)}>
          {required && <span className="text-danger">* </span>}
          {label}
        </span>
      )}
      <div className="relative">
        {leftIcon && (
          <span className="absolute left-3 top-1/2 -translate-y-1/2">
            {leftIcon}
          </span>
        )}
        <input
          type={type}
          required={required}
          value={value}
          className={cn(inputClasses, {
            "pl-10": leftIcon,
            "pr-10": rightIcon,
          })}
          ref={ref}
          {...rest}
        />
        {rightIcon && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2">
            {rightIcon}
          </span>
        )}
      </div>
      {helpText && (
        <span
          className={cn(
            "text-[0.9rem] font-medium inline-block left-[2px] mt-1 text-gray-300",
            {
              "text-danger": hasError,
            }
          )}
        >
          {helpText}
        </span>
      )}
    </label>
  );
});

Input.displayName = "Input";
