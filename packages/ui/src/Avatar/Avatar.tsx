import { VariantProps, cva } from 'cva';
import { cn } from 'lib';
import { FC, HTMLAttributes } from 'react';
import { TbUser } from 'react-icons/tb';

const avatar = cva(
  'inline-flex justify-center items-center aspect-square overflow-hidden text-center font-medium',
  {
    variants: {
      rounded: {
        full: 'rounded-full',
        none: 'rounded-none',
        sm: 'rounded-sm',
        md: 'rounded-md',
        lg: 'rounded-lg',
      },
      color: {
        primary: 'text-white bg-primary',
        secondary: 'text-white bg-secondary',
        gray: 'text-black bg-gray',
      },
      size: {
        sm: 'h-7 text-sm leading-7',
        md: 'h-9 text-base leading-9',
        lg: 'h-12 text-lg leading-[3rem]',
      },
    },
    defaultVariants: {
      size: 'md',
      rounded: 'full',
      color: 'primary',
    },
  }
);

export interface AvatarProps
  extends Omit<HTMLAttributes<HTMLDivElement>, 'color'>,
    VariantProps<typeof avatar> {
  src?: string;
  name?: string;
}

const getInitials = (name: string) => {
  if (!name) {
    return 'U';
  }

  const names = name.split(' ');
  let initials = '';

  initials += names[0][0]; // First initial of the first name

  if (names.length > 1) {
    initials += names[names.length - 1][0]; // First initial of the last name
  }

  return initials.toUpperCase();
};

export const Avatar: FC<AvatarProps> = (props) => {
  const { className, src, name, color, size, rounded, ...rest } = props;
  const classes = cn(avatar({ rounded, color, size }), className);

  return (
    <div className={classes} {...rest}>
      {src && (
        <img
          src={src}
          alt={name}
          className={cn('w-full h-auto object-cover', {
            'rounded-full': rounded === 'full',
            'rounded-sm': rounded === 'sm',
            'rounded-md': rounded === 'md',
            'rounded-lg': rounded === 'lg',
          })}
        />
      )}
      {!src && name && <span title={name}>{getInitials(name)}</span>}
      {!src && !name && (
        <TbUser
          className={cn('h-6 w-auto', {
            'h-5': size === 'sm',
          })}
        />
      )}
    </div>
  );
};

Avatar.displayName = 'Avatar';
