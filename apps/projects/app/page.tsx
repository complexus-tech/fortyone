import { BsFillCheckCircleFill } from 'react-icons/bs';
import { TbAdjustmentsHorizontal, TbPlus } from 'react-icons/tb';
import { WiMoonAltThirdQuarter } from 'react-icons/wi';
import { Box, BreadCrumbs, Button, Container, Flex, Text } from 'ui';
import { Sidebar } from './components/shared';
import { Issue } from './components/ui';

export default function Page(): JSX.Element {
  return (
    <main>
      <Box className='grid grid-cols-[220px_auto] h-screen'>
        <Sidebar />
        <Box>
          <Container className='border-b dark:border-dark-100 border-gray-100 h-16 flex items-center'>
            <Flex className='w-full' justify='between'>
              <BreadCrumbs
                breadCrumbs={[{ name: 'projects' }, { name: 'test' }]}
              />

              <Flex gap={2}>
                <Button size='sm' leftIcon={<TbPlus className='h-5 w-auto' />}>
                  New issue
                </Button>
                <Button
                  leftIcon={<TbAdjustmentsHorizontal className='h-5 w-auto' />}
                  color='tertiary'
                  size='sm'
                  variant='outline'
                >
                  Display
                </Button>
              </Flex>
            </Flex>
          </Container>
          <Box className='h-[calc(100vh-4rem)] pb-3'>
            <Container className='bg-gray-50 dark:bg-dark-200 py-3'>
              <Flex justify='between' align='center'>
                <Flex align='center' gap={2}>
                  <WiMoonAltThirdQuarter className='h-5 text-warning w-auto' />
                  <Text fontWeight='medium'>In Progress</Text>
                  <Text color='muted'>12</Text>
                </Flex>
                <TbPlus className='h-5 w-auto dark:text-gray-200' />
              </Flex>
            </Container>
            {new Array(10).fill(0).map((_, i) => (
              <Issue key={i} />
            ))}
            <Container className='bg-gray-50 dark:bg-dark-200 py-3'>
              <Flex justify='between' align='center'>
                <Flex align='center' gap={2}>
                  <BsFillCheckCircleFill className='h-4 text-success w-auto' />
                  <Text fontWeight='medium'>Done</Text>
                  <Text color='muted'>5</Text>
                </Flex>

                <TbPlus className='h-5 w-auto dark:text-gray-200' />
              </Flex>
            </Container>
            {new Array(4).fill(0).map((_, i) => (
              <Issue key={i} />
            ))}
          </Box>
        </Box>
      </Box>
    </main>
  );
}
