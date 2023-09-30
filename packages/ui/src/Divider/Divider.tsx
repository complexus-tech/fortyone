import { cn } from 'lib';

export const Divider = ({ className = '' }) => {
  return (
    <div
      className={cn('border-gray-100 dark:border-gray-300 border-t', className)}
    />
  );
};
