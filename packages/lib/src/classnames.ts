import classes, { ArgumentArray } from 'classnames';
import { twMerge } from 'tailwind-merge';

export const cn = (...args: ArgumentArray) => {
  return twMerge(classes(args));
};
