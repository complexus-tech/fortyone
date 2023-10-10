import { TbAdjustmentsHorizontal } from 'react-icons/tb';
import { BreadCrumbs, Button, Flex } from 'ui';
import { BodyContainer, HeaderContainer } from '../components/shared';
import {
  Issue,
  IssueHeader,
  IssuesToolbar,
  NewIssueButton,
} from '../components/ui';

export default function Page(): JSX.Element {
  return (
    <>
      <HeaderContainer className='justify-between'>
        <BreadCrumbs
          breadCrumbs={[{ name: 'My issues' }, { name: 'Active' }]}
        />
        <Flex gap={2}>
          <NewIssueButton />
          <Button
            leftIcon={<TbAdjustmentsHorizontal className='h-5 w-auto' />}
            color='tertiary'
            size='sm'
            variant='outline'
          >
            Display
          </Button>
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <IssueHeader count={3} status='Backlog' />
        {new Array(3).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='Backlog'
            title='These issues are not assigned to any sprint.'
          />
        ))}
        <IssueHeader count={2} status='Todo' />
        {new Array(2).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='Todo'
            title='These issues are at the top of the backlog and are ready to be worked on.'
          />
        ))}
        <IssueHeader count={3} status='In Progress' />
        {new Array(3).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='In Progress'
            title='These issues are being actively worked on.'
          />
        ))}
        <IssueHeader count={3} status='Testing' />
        {new Array(3).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='Testing'
            title='These issues are being tested by the QA team.'
          />
        ))}
        <IssueHeader count={4} status='Done' />
        {new Array(4).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='Done'
            title='These issues are completed and ready to be deployed.'
          />
        ))}
        <IssueHeader count={2} status='Canceled' />
        {new Array(2).fill(0).map((_, i) => (
          <Issue
            key={i}
            status='Canceled'
            title='These issues are no longer being worked on.'
          />
        ))}

        <IssuesToolbar />
      </BodyContainer>
    </>
  );
}
