import { Box, Container, Divider, Text, TextEditor } from "ui";
import {
  Activities,
  Attachments,
  Reactions,
  SubissuesButton,
} from "@/app/issue/components";

export const MainDetails = () => {
  return (
    <Box className="h-full overflow-y-auto border-r border-gray-50 pb-8 dark:border-dark-200">
      <Container className="pt-6">
        <Text className="mb-6" fontSize="3xl" fontWeight="medium">
          Change the color of the button to red
        </Text>
        <TextEditor
          content="
          <h3>Jira Issue Description Example</h3>
          <p>This is a sample HTML text for a Jira issue description. It can include various elements such as headings, paragraphs, lists, and more.</p>
          <h2>Steps to Reproduce:</h2>
          <ol>
              <li>Open the application.</li>
              <li>Go to the settings page.</li>
              <li>Change the language to French.</li>
              <li>Save the settings.</li>
          </ol>
            "
        />
        <Reactions />
        <SubissuesButton />
        <Divider className="my-4" />
        <Attachments />
        <Divider className="mb-6 mt-8" />
        <Activities />
      </Container>
    </Box>
  );
};
