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
        <span className='mb-3 block font-medium dark:text-white'>
          {label}
          {required && <span className='text-danger'>*</span>}
        </span>
      )}
      <input
        type={type}
        required={required}
        value={value}
        className={cn(
          'w-full rounded-xl border-[1.5px] dark:bg-transparent border-gray-100 dark:border-gray-300 dark:ring-offset-blue-dark py-[0.9rem] px-5 placeholder:text-gray-200 read-only:cursor-default focus:border-gray-100 dark:focus:border-gray-300 focus:outline-0 focus:ring focus:ring-primary focus:ring-offset-2 read-only:focus:ring-0 dark:text-white',
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
