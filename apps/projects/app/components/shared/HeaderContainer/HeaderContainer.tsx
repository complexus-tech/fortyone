import { cn } from 'lib';
import { Container, ContainerProps } from 'ui';

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        'border-b dark:border-dark-100 border-gray-100 h-16 flex items-center',
        className
      )}
    >
      {children}
    </Container>
  );
};
