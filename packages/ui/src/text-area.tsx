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
        <span className="mb-[0.35rem] block text-foreground font-medium">
          {label}
          {required && <span className="text-danger">*</span>}
        </span>
      )}
      <textarea
        required={required}
        className={cn(
          "w-full rounded-[0.6rem] border bg-surface border-input px-4 leading-[2.8rem] focus-visible:outline-0 focus-visible:ring-2 focus-visible:ring-ring read-only:focus-visible:ring-0 placeholder:text-text-muted",
          className
        )}
        {...rest}
      />
    </label>
  );
};
