import { cn } from 'lib';
import { FC, InputHTMLAttributes, useId } from 'react';

export interface InputButtonProps
  extends InputHTMLAttributes<HTMLInputElement> {
  className?: string;
  active?: boolean;
  label?: string;
  type?: 'checkbox' | 'radio';
}

export const InputButton: FC<InputButtonProps> = (props) => {
  const uid = useId();
  const { children, label, className, id, type = 'checkbox', ...rest } = props;
  return (
    <span className='[&_input:checked_+_label]:text-black [&_input:checked_+_label]:bg-primary [&_input:checked_+_label]:border-primary'>
      {label && <span className='mb-3 block dark:text-white'>{label}</span>}
      <input
        className='invisible absolute opacity-0'
        id={id || uid}
        type={type}
        {...rest}
      />
      <label
        htmlFor={id || uid}
        className={cn(
          'mb-1 inline-block cursor-pointer rounded-xl border-[1.5px] border-gray-100 dark:border-gray-300 px-6 py-[0.9rem] transition duration-200 ease-linear dark:text-white font-medium hover:border-primary hover:bg-primary hover:bg-opacity-10 hover:opacity-100',
          className
        )}
      >
        {children}
      </label>
    </span>
  );
};
