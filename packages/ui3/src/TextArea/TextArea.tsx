import { cn } from "lib";
import { FC, TextareaHTMLAttributes } from "react";

interface Props extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
}

export const TextArea: FC<Props> = (props) => {
  const { label, required, className, ...rest } = props;
  return (
    <label>
      {label && (
        <span className="mb-[0.35rem] block dark:text-white font-medium">
          {label}
          {required && <span className="text-danger">*</span>}
        </span>
      )}
      <textarea
        required={required}
        className={cn(
          "w-full rounded-[0.6rem] border bg-white/70 dark:bg-dark/20 border-gray-100 dark:border-dark-100 dark:ring-offset-dark px-4 leading-[2.8rem] focus:outline-0 focus:ring-[2.5px] focus:ring-gray-100 dark:focus:ring-dark-50 focus:ring-offset-1 read-only:focus:ring-0 placeholder:text-gray/80 dark:placeholder:text-gray-300",
          className
        )}
        {...rest}
      />
    </label>
  );
};
