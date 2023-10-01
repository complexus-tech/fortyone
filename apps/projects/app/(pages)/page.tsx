import { TbAdjustmentsHorizontal } from 'react-icons/tb';
import { BreadCrumbs, Button, Flex } from 'ui';
import { BodyContainer, HeaderContainer } from '../components/shared';
import { Issue, IssueHeader, NewIssueButton } from '../components/ui';

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
          <Issue key={i} status='Backlog' />
        ))}
        <IssueHeader count={3} status='In Progress' />
        {new Array(3).fill(0).map((_, i) => (
          <Issue key={i} status='In Progress' />
        ))}
        <IssueHeader count={3} status='Testing' />
        {new Array(3).fill(0).map((_, i) => (
          <Issue key={i} status='Testing' />
        ))}
        <IssueHeader count={4} status='Done' />
        {new Array(4).fill(0).map((_, i) => (
          <Issue key={i} status='Done' />
        ))}
        <IssueHeader count={2} status='Canceled' />
        {new Array(2).fill(0).map((_, i) => (
          <Issue key={i} status='Canceled' />
        ))}
      </BodyContainer>
    </>
  );
}
