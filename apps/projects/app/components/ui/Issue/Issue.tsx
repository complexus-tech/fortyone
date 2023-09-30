import { GoDotFill } from 'react-icons/go';
import { TbDots } from 'react-icons/tb';
import { WiMoonAltThirdQuarter } from 'react-icons/wi';
import { Avatar, Button, Container, Flex, Text } from 'ui';

export const Issue = () => {
  return (
    <Container className='group border-b hover:bg-gray-50/50 hover:dark:bg-dark-200/50 py-3 dark:border-dark-200 border-gray-50 flex items-center justify-between'>
      <Flex align='center' gap={2} className='relative'>
        <input
          defaultChecked
          type='checkbox'
          className='accent-primary h-5 aspect-square absolute -left-6 hidden group-hover:inline-block'
        />
        <TbDots />
        <Text color='muted' className='w-[60px] truncate'>
          COM-12
        </Text>
        <WiMoonAltThirdQuarter className='h-5 text-warning w-auto' />
        <Text
          fontWeight='medium'
          className='whitespace-nowrap text-ellipsis overflow-hidden w-[70vh]'
        >
          The quick brown fox jumps over the lazy dog. The quick brown fox jumps
          over the lazy dog. The quick brown fox jumps over the lazy dog. The
          quick brown fox jumps over the lazy dog.
        </Text>
      </Flex>
      <Flex align='center' gap={3}>
        <Flex align='center' gap={1}>
          <Button
            size='xs'
            rounded='xl'
            color='tertiary'
            variant='outline'
            leftIcon={<GoDotFill className='text-info' />}
          >
            Feature
          </Button>
          <Button
            size='xs'
            rounded='xl'
            color='tertiary'
            variant='outline'
            leftIcon={<GoDotFill className='text-danger' />}
          >
            Bug
          </Button>
        </Flex>

        <Text color='muted'>Sep 27</Text>
        <Avatar
          name='Joseph Mukorivo'
          size='sm'
          src='https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo'
        />
      </Flex>
    </Container>
  );
};
