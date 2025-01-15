import { FC, InputHTMLAttributes } from "react";

import { cn } from "lib";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  helpText?: string;
}

export const Input: FC<InputProps> = (props) => {
  const {
    label,
    className,
    required,
    value,
    helpText,
    type = "text",
    ...rest
  } = props;
  return (
    <label className="relative block">
      {label && (
        <span className="mb-[0.35rem] block">
          {label}
          {required && <span className="text-danger">*</span>}
        </span>
      )}
      <input
        type={type}
        required={required}
        value={value}
        className={cn(
          "w-full rounded-[0.45rem] border bg-white/70 dark:bg-dark/20 border-gray-100 dark:border-dark-100 dark:ring-offset-dark px-4 h-[2.8rem] leading-[2.8rem] focus:outline-0 focus:ring-[2.5px] focus:ring-gray-100 dark:focus:ring-dark-50 focus:ring-offset-1 read-only:focus:ring-0 placeholder:text-gray/80 dark:placeholder:text-gray-300",
          className
        )}
        {...rest}
      />
      {helpText && (
        <span className="text-[0.9rem] font-medium inline-block left-[2px] mt-1 text-gray-300">
          {helpText}
        </span>
      )}
    </label>
  );
};
