'use client';
import { Box } from 'ui';
import { Header } from './Header';
import { Navigation } from './Navigation';
import { Projects } from './Projects';

export const Sidebar = () => {
  return (
    <Box className='border-r border-gray-100 dark:border-dark-100 h-screen px-4'>
      <Header />
      <Navigation />
      <Projects />
    </Box>
  );
};
