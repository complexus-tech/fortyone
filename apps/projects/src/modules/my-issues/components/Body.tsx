import { BodyContainer } from '../../../components/shared';
import { Issue, IssueHeader, IssuesToolbar } from '../../../components/ui';

export const Body = () => {
  return (
    <BodyContainer className='relative'>
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
          priority='High'
          status='Todo'
          title='These issues are at the top of the backlog and are ready to be worked on.'
        />
      ))}
      <IssueHeader count={3} status='In Progress' />
      {new Array(3).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='In Progress'
          priority='Urgent'
          title='These issues are being actively worked on.'
        />
      ))}
      <IssueHeader count={3} status='Testing' />
      {new Array(3).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Testing'
          priority='Medium'
          title='These issues are being tested by the QA team.'
        />
      ))}
      <IssueHeader count={4} status='Done' />
      {new Array(4).fill(0).map((_, i) => (
        <Issue
          key={i}
          status='Done'
          priority='Low'
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
  );
};
