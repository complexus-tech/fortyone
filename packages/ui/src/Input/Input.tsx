import { FC, InputHTMLAttributes } from 'react';

import { cn } from 'lib';

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
    type = 'text',
    ...rest
  } = props;
  return (
    <label className='relative block'>
      {label && (
        <span className='mb-2 block'>
          {label}
          {required && <span className='text-danger'>*</span>}
        </span>
      )}
      <input
        type={type}
        required={required}
        value={value}
        className={cn(
          'w-full rounded-lg border dark:bg-transparent border-gray-100 dark:border-dark-100 dark:ring-offset-dark px-4 h-[2.5rem] leading-[2.5rem] focus:outline-0 focus:ring-[1.5px] focus:ring-primary focus:ring-offset-2 read-only:focus:ring-0',
          className
        )}
        {...rest}
      />
      {helpText && (
        <span className='text-[0.8rem] font-medium inline-block left-[2px] -bottom-5 absolute text-gray-300 first-letter:uppercase'>
          {helpText}
        </span>
      )}
    </label>
  );
};
