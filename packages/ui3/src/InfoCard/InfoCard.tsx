import { cn } from 'lib';
import { Box } from '../Box/Box';
import { Text } from '../Text/Text';

export const InfoCard = ({
  property,
  className = '',
  value,
  small,
  raw = false,
}: {
  property: string;
  value: any;
  raw?: boolean;
  small?: boolean;
  className?: string;
}) => (
  <Box className='mb-3 grid grid-cols-6'>
    <Text className='col-span-2'>
      <span className='opacity-90'>{property}:</span>
    </Text>
    <Text
      as='span'
      color='muted'
      fontWeight='bold'
      className={cn(
        'col-span-4 max-w-[360px] truncate',
        {
          lowercase: small && !raw,
          capitalize: !small && !raw,
        },
        className
      )}
    >
      {value}
    </Text>
  </Box>
);
