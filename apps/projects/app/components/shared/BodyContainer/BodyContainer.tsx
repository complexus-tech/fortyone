import { cn } from 'lib';
import { Box, BoxProps } from 'ui';

export const BodyContainer = ({ children, className }: BoxProps) => {
  return (
    <Box className={cn('h-[calc(100vh-4rem)] pb-3 overflow-y-auto', className)}>
      {children}
    </Box>
  );
};
