import { cn } from 'lib';
import Image from 'next/image';
import { useState } from 'react';

import { Box } from '../Box/Box';

export const BlurImage = ({
  className = '',
  rounded = false,
  priority = false,
  objectPosition = 'center',
  quality = 75,
  theme = 'light',
  alt = '',
  src = '',
}) => {
  const [isLoading, setLoading] = useState(true);

  return (
    <Box
      className={cn(
        'group relative w-full overflow-hidden',
        {
          'bg-gray-100': theme === 'light',
          'bg-[#000]': theme === 'dark',
          'rounded-xl lg:rounded-3xl': rounded,
        },
        className
      )}
    >
      <Image
        alt={alt}
        src={src}
        priority={priority}
        quality={quality}
        layout='fill'
        objectPosition={objectPosition}
        objectFit='cover'
        className={cn(
          'border-separate scale-100 blur-0 grayscale-0 transition duration-500 ease-in-out',
          {
            'rounded-xl lg:rounded-3xl': rounded,
            'scale-110 blur-2xl grayscale': isLoading,
          }
        )}
        onLoadingComplete={() => setLoading(false)}
      />
    </Box>
  );
};
