import { cn } from 'lib';
import { Container, ContainerProps } from 'ui';

export const HeaderContainer = ({ children, className }: ContainerProps) => {
  return (
    <Container
      className={cn(
        'border-b z-10 dark:border-dark-100 border-gray-100 h-16 fixed w-[calc(100vw-220px)] right-0 top-0 bg-white/50 dark:bg-dark/50 backdrop-blur flex items-center',
        className
      )}
    >
      {children}
    </Container>
  );
};
