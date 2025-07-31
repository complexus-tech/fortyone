import { Box } from "ui";
import { Hero } from "./components/hero";
import { Content } from "./components/content";
import { Form } from "./components/form";

export const SubmitStoryPage = () => {
  return (
    <Box>
      <Hero />
      <Content />
      <Form />
    </Box>
  );
};
