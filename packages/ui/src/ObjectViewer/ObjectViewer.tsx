import { cn } from 'lib';

export const ObjectViewer = ({
  data,
  type,
  className = '',
  onError,
}: {
  data: string;
  type: string;
  className?: string;
  onError?: (e: any) => void;
}) => {
  return (
    <object
      type={type}
      data={data}
      className={cn('block w-full h-full', className)}
    />
  );
};
