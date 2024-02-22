import { ArrowDownIcon } from "icons";
import { cn } from "lib";
import { FC, OptionHTMLAttributes, SelectHTMLAttributes } from "react";

interface Props extends SelectHTMLAttributes<HTMLSelectElement> {
  className?: string;
  label?: string;
}

interface OptionProps extends OptionHTMLAttributes<HTMLOptionElement> {
  className?: string;
}

interface StaticComponents {
  Option?: FC<OptionProps>;
}

const Option: FC<OptionProps> = (props) => {
  const { className, children, ...rest } = props;
  return (
    <option className={className} {...rest}>
      {children}
    </option>
  );
};

export const Select: FC<Props> & StaticComponents = (props) => {
  const { className, label, required, children, ...rest } = props;
  return (
    <label>
      {label && (
        <span className="mb-3 block">
          {label}
          {required && <span className="text-danger">*</span>}
        </span>
      )}
      <div className="relative flex items-center">
        <select
          required={required}
          className={cn(
            "w-full appearance-none rounded-lg border border-gray-100 py-3 px-5 placeholder:text-gray-200 focus:border-gray-100 focus:outline-0 focus:ring focus:ring-primary focus:ring-offset-2 2xl:py-4",
            className
          )}
          {...rest}
        >
          {children}
        </select>
        <ArrowDownIcon className="pointer-events-none absolute right-3 h-4 w-auto" />
      </div>
    </label>
  );
};

Select.Option = Option;
