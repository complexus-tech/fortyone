import { DeleteIcon } from "icons";
import { Divider, Text, Button, Box } from "ui";

export default function Page(): JSX.Element {
  return (
    <Box>
      <Text className="mb-3" fontSize="3xl">
        Delete workspace
      </Text>
      <Text color="muted">
        Deleting a workspace is irreversible. Please be certain. This action
        will delete all data associated with the workspace.
      </Text>
      <Divider className="my-5" />
      <Button
        leftIcon={<DeleteIcon className="h-[1.15rem] w-auto" />}
        rounded="sm"
      >
        Delete workspace
      </Button>
    </Box>
  );
}
