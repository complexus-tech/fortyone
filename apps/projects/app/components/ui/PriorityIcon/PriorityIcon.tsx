import { BsFillExclamationSquareFill } from 'react-icons/bs';
import {
  TbAntennaBars1,
  TbAntennaBars2,
  TbAntennaBars3,
  TbAntennaBars4,
} from 'react-icons/tb';

export const PriorityIcon = ({
  priority = 'No Priority',
}: {
  priority?: 'No Priority' | 'Urgent' | 'High' | 'Medium' | 'Low';
}) => {
  return (
    <>
      {priority === 'No Priority' && (
        <TbAntennaBars1 strokeWidth={2.5} className='h-6 text-gray w-auto' />
      )}
      {priority === 'Urgent' && (
        <BsFillExclamationSquareFill className='h-[1.1rem] text-danger w-auto' />
      )}
      {priority === 'High' && (
        <TbAntennaBars4 strokeWidth={2.5} className='h-6 text-gray w-auto' />
      )}
      {priority === 'Medium' && (
        <TbAntennaBars3 strokeWidth={2.5} className='h-6 text-gray w-auto' />
      )}
      {priority === 'Low' && (
        <TbAntennaBars2 strokeWidth={2.5} className='h-6 text-gray w-auto' />
      )}
    </>
  );
};
