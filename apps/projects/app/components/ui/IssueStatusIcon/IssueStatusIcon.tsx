import {
  TbBrandParsinta,
  TbCircleCheckFilled,
  TbCircleDashed,
  TbCircleXFilled,
  TbProgress,
  TbWashDryclean,
} from 'react-icons/tb';
export const IssueStatusIcon = ({
  status = 'Backlog',
}: {
  status?:
    | 'Backlog'
    | 'Todo'
    | 'In Progress'
    | 'Testing'
    | 'Done'
    | 'Duplicate'
    | 'Canceled';
}) => {
  return (
    <>
      {status === 'Backlog' && (
        <TbCircleDashed className='h-5 text-gray w-auto' />
      )}
      {status === 'Todo' && (
        <TbWashDryclean className='h-5 text-gray/60 w-auto' />
      )}
      {status === 'In Progress' && (
        <TbProgress className='h-5 text-warning w-auto' />
      )}
      {status === 'Testing' && (
        <TbBrandParsinta className='h-5 text-info w-auto' />
      )}
      {status === 'Done' && (
        <TbCircleCheckFilled className='h-5 text-success w-auto' />
      )}
      {status === 'Canceled' && (
        <TbCircleXFilled className='h-5 text-danger w-auto' />
      )}
      {status === 'Duplicate' && (
        <TbCircleXFilled className='h-5 text-warning w-auto' />
      )}
    </>
  );
};
