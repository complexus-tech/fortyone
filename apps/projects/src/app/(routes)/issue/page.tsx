import { TbClipboardCopy, TbLink, TbMoodPlus, TbTrash } from 'react-icons/tb';
import {
  Badge,
  Box,
  BreadCrumbs,
  Button,
  Container,
  Divider,
  Flex,
  Text,
  Tooltip,
} from 'ui';
import { BodyContainer, HeaderContainer } from '../../../components/shared';
import {
  IssueStatusIcon,
  NewIssueButton,
  PriorityIcon,
} from '../../../components/ui';

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer>
        <Flex justify='between' align='center' className='w-full'>
          <BreadCrumbs
            breadCrumbs={[
              { name: 'Complexus' },
              { name: 'Web design' },
              { name: 'COM-12' },
            ]}
          />
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer className='grid grid-cols-[auto_400px]'>
        <Box className='border-r dark:border-dark-100 border-gray-100'>
          <Container className='pt-6'>
            <Text fontSize='3xl' className='mb-6'>
              This should be an issue title
            </Text>
            <Text color='muted' fontSize='lg' className='leading-7'>
              This will hold the description of the issue. It will be a long one
              so we can test the overflow of the container. This will hold the
              description of the issue. It will be a long one. The quick brown
              fox jumped over the lazy dog. This will hold the description of
              the issue. It will be a long one. The quick brown fox jumped over
              the lazy dog. This will hold the description of the issue.
            </Text>
            <Flex gap={1} className='mt-4'>
              <Badge
                rounded='full'
                size='lg'
                color='tertiary'
                variant='outline'
              >
                👍🏼 1
              </Badge>
              <Badge
                rounded='full'
                size='lg'
                color='tertiary'
                variant='outline'
              >
                🇿🇼 3
              </Badge>
              <Badge
                rounded='full'
                size='lg'
                color='tertiary'
                variant='outline'
              >
                👌 2
              </Badge>
              <Badge
                rounded='full'
                size='lg'
                className='px-2'
                color='tertiary'
                variant='outline'
              >
                <TbMoodPlus className='h-5 w-auto opacity-80' />
              </Badge>
            </Flex>
          </Container>
        </Box>
        <Box className='bg-gray-50/30 dark:bg-dark-200/40'>
          <Box className='h-16 border-b dark:border-dark-100 flex items-center border-gray-100'>
            <Container className='w-full flex justify-between items-center'>
              <Text fontWeight='medium' color='muted'>
                COMP-13
              </Text>
              <Flex gap={2}>
                <Tooltip title='Copy issue link'>
                  <Button
                    variant='naked'
                    color='tertiary'
                    leftIcon={<TbLink className='h-5 w-auto' />}
                  >
                    <span className='sr-only'>Copy issue link</span>
                  </Button>
                </Tooltip>
                <Tooltip title='Copy issue id'>
                  <Button
                    variant='naked'
                    color='tertiary'
                    leftIcon={<TbClipboardCopy className='h-5 w-auto' />}
                  >
                    <span className='sr-only'>Copy issue id</span>
                  </Button>
                </Tooltip>
                <Tooltip title='Delete issue'>
                  <Button
                    variant='naked'
                    color='danger'
                    leftIcon={<TbTrash className='h-5 w-auto' />}
                  >
                    <span className='sr-only'>Delete issue</span>
                  </Button>
                </Tooltip>
              </Flex>
            </Container>
          </Box>
          <Container className='pt-6'>
            <Box className='grid grid-cols-[8rem_auto] gap-3 items-center my-4'>
              <Text className='truncate' fontWeight='medium' color='muted'>
                Status
              </Text>
              <Button
                size='lg'
                className='px-5'
                color='tertiary'
                variant='naked'
              >
                <Box className='grid grid-cols-[1.5rem_auto] gap-1 items-center'>
                  <IssueStatusIcon status='In Progress' />
                  <Text>In Progress</Text>
                </Box>
              </Button>
            </Box>
            <Box className='grid grid-cols-[8rem_auto] gap-3 items-center my-4'>
              <Text className='truncate' fontWeight='medium' color='muted'>
                Priority
              </Text>
              <Button
                size='lg'
                className='px-5'
                color='tertiary'
                variant='naked'
              >
                <Box className='grid grid-cols-[1.5rem_auto] gap-1 items-center'>
                  <PriorityIcon priority='High' />
                  <Text>High</Text>
                </Box>
              </Button>
            </Box>

            <Box className='grid grid-cols-[8rem_auto] gap-3 items-center my-4'>
              <Text className='truncate' fontWeight='medium' color='muted'>
                Assignee
              </Text>
              <Button
                size='lg'
                className='px-5'
                color='tertiary'
                variant='naked'
              >
                <Box className='grid grid-cols-[1.5rem_auto] gap-1 items-center'>
                  <IssueStatusIcon status='Backlog' />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
            <Box className='grid grid-cols-[8rem_auto] gap-3 items-center my-4'>
              <Text className='truncate' fontWeight='medium' color='muted'>
                Assignee
              </Text>
              <Button
                size='lg'
                className='px-5'
                color='tertiary'
                variant='naked'
              >
                <Box className='grid grid-cols-[1.5rem_auto] gap-1 items-center'>
                  <IssueStatusIcon status='Done' />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
            <Divider />
            <Box className='grid grid-cols-[8rem_auto] gap-3 items-center my-4'>
              <Text className='truncate' fontWeight='medium' color='muted'>
                Labels
              </Text>
              <Button
                size='lg'
                className='px-5'
                color='tertiary'
                variant='naked'
              >
                <Box className='grid grid-cols-[1.5rem_auto] gap-1 items-center'>
                  <IssueStatusIcon status='Backlog' />
                  <Text>Backlog</Text>
                </Box>
              </Button>
            </Box>
          </Container>
        </Box>
      </BodyContainer>
    </>
  );
}
