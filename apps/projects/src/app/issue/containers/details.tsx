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
            <h1>Test heading</h1>
            <p>Test paragraph</p>
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
