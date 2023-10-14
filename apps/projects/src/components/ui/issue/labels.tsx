import { GoDotFill } from 'react-icons/go';
import { Button, Flex } from 'ui';

export const Labels = () => {
  return (
    <Flex align='center' gap={1}>
      <Button
        color='tertiary'
        leftIcon={<GoDotFill className='text-info' />}
        rounded='xl'
        size='xs'
        variant='outline'
      >
        Feature
      </Button>
      <Button
        color='tertiary'
        leftIcon={<GoDotFill className='text-danger' />}
        rounded='xl'
        size='xs'
        variant='outline'
      >
        Bug
      </Button>
    </Flex>
  );
};
