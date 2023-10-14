import { RiDraggable } from 'react-icons/ri';

export const DragHandle = () => {
  return (
    <RiDraggable className='absolute h-5 w-auto -left-[2.9rem] text-gray cursor-move hidden group-hover:inline-block' />
  );
};
