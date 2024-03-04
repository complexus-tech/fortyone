import { DeleteIcon } from "icons";
import { Divider, Text, Avatar, Input, Button } from "ui";

export default function Page(): JSX.Element {
  return (
    <div>
      <Text className="mb-3" fontSize="3xl">
        Workspace settings
      </Text>
      <Text color="muted">Manage your workspace settings.</Text>
      <Divider className="my-5" />
      <Text className="mb-2" fontSize="lg" fontWeight="medium">
        Logo
      </Text>
      <Avatar
        className="mb-5 h-28"
        name="Joseph Mukorivo"
        rounded="lg"
        src="/complexus.png"
      />
      <Text className="mb-5" color="muted">
        Choose a logo for your workspace.
      </Text>

      <Input className="mb-4" label="Workspace name" required value="Joseph" />
      <Input className="mb-4" label="Workspace url" required value="Joseph" />
      <Input className="mb-4" label="Company size" required value="Joseph" />
      <Button rounded="sm">Update workspace</Button>

      <Divider className="my-5" />

      <Text className="mb-3" fontSize="xl">
        Danger zone
      </Text>

      <Text className="mb-5" color="muted">
        Deleting a workspace is irreversible. Please be certain. This action
        will delete all data associated with the workspace.
      </Text>

      <Button
        leftIcon={<DeleteIcon className="h-[1.15rem] w-auto" />}
        rounded="sm"
      >
        Delete workspace
      </Button>
    </div>
  );
}
