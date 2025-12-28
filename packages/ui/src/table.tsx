import { VariantProps, cva } from 'cva';
import { cn } from 'lib';
import { HTMLAttributes, TdHTMLAttributes, ThHTMLAttributes } from 'react';

const table = cva('w-full border-collapse relative dark:text-tertiary', {
  variants: {
    size: {
      xs: '[&_td]:py-[0.2rem]',
      sm: '[&_td]:py-[0.4rem]',
      md: '[&_td]:py-[0.7rem]',
      lg: '[&_td]:py-4',
    },
    align: {
      left: '[&_td]:text-left [&_th]:text-left first:[&_td]:pl-[2px] [&_td]:pr-0',
      right:
        '[&_td]:text-right [&_td]:pl-2 [&_td]:pr-0 [&_th]:text-right first:[&_th]:text-left first:[&_td]:text-left first:[&_td]:pl-0',
    },
    color: {
      primary: 'hover:[&_tbody_tr]:bg-primary/20',
      secondary: 'hover:[&_tbody_tr]:bg-secondary/20',
      light: 'hover:[&_tbody_tr]:bg-tertiary/20',
      info: 'hover:[&_tbody_tr]:bg-info/20',
      danger: 'hover:[&_tbody_tr]:bg-danger/20',
      warning: 'hover:[&_tbody_tr]:bg-warning/20',
    },
    variant: {
      striped: null,
      bordered: null,
    },
  },
  compoundVariants: [
    {
      variant: 'striped',
      color: 'primary',
      className: 'odd:[&_tbody_tr]:bg-primary/12',
    },
    {
      variant: 'striped',
      color: 'secondary',
      className: 'odd:[&_tbody_tr]:bg-secondary/12',
    },
    {
      variant: 'striped',
      color: 'light',
      className: 'odd:[&_tbody_tr]:bg-tertiary/[0.12]',
    },
    {
      variant: 'striped',
      color: 'info',
      className: 'odd:[&_tbody_tr]:bg-info/12',
    },
    {
      variant: 'striped',
      color: 'danger',
      className: 'odd:[&_tbody_tr]:bg-danger/12',
    },
    {
      variant: 'striped',
      color: 'warning',
      className: 'odd:[&_tbody_tr]:bg-warning/12',
    },
    {
      variant: 'bordered',
      color: ['primary', 'secondary', 'light', 'info', 'danger', 'warning'],
      className: '[&_tr]:border-b',
    },
  ],
  defaultVariants: {
    size: 'md',
    align: 'left',
    variant: 'striped',
    color: 'primary',
  },
});

type TableProps = HTMLAttributes<HTMLTableElement> & VariantProps<typeof table>;
export const Table = ({
  children,
  className,
  align,
  size,
  color,
  variant,
  ...rest
}: TableProps) => {
  const classes = cn(table({ align, size, color, variant }), className);
  return (
    <table className={classes} {...rest}>
      {children}
    </table>
  );
};

type HeadProps = HTMLAttributes<HTMLTableSectionElement>;
const Head = ({ children, className, ...rest }: HeadProps) => {
  return (
    <thead
      className={cn(
        'w-full border-b border-gray-100 dark:border-gray-300',
        className
      )}
      {...rest}
    >
      {children}
    </thead>
  );
};

type ThProps = ThHTMLAttributes<HTMLTableCellElement>;
const Th = ({ children, className, ...rest }: ThProps) => {
  return (
    <th
      className={cn(
        'bg-white dark:bg-dark-200 z-10 min-w-[25px] py-3 sticky top-0 text-[0.88rem] font-semibold capitalize text-gray-300 dark:text-white',
        className
      )}
      {...rest}
    >
      {children}
    </th>
  );
};

type BodyProps = HTMLAttributes<HTMLTableSectionElement>;
const Body = ({ children, className, ...rest }: BodyProps) => {
  return (
    <tbody className={className} {...rest}>
      {children}
    </tbody>
  );
};

type TdProps = TdHTMLAttributes<HTMLTableCellElement>;
const Td = ({ children, className, ...rest }: TdProps) => {
  return (
    <td
      className={cn('first:font-semibold pr-2 last:pr-0', className)}
      {...rest}
    >
      {children}
    </td>
  );
};

type TrProps = HTMLAttributes<HTMLTableRowElement>;
const Tr = ({ children, className, ...rest }: TrProps) => {
  return (
    <tr
      className={cn(
        'border-tertiary-50 dark:border-dark-50 last:border-b-0 transition duration-200 ease-linear',
        className
      )}
      {...rest}
    >
      {children}
    </tr>
  );
};

Table.Head = Head;
Table.Body = Body;
Table.Th = Th;
Table.Tr = Tr;
Table.Td = Td;

Table.displayName = 'Table';
Head.displayName = 'Table.Head';
Body.displayName = 'Table.Body';
Th.displayName = 'Table.Th';
Tr.displayName = 'Table.Tr';
Td.displayName = 'Table.Td';
