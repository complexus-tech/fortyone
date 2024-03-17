import { Divider, Text, Avatar, Input, Button, Box } from "ui";

export default function Page(): JSX.Element {
  return (
    <Box>
      <Text className="mb-3" fontSize="3xl">
        Workspace settings
      </Text>
      <Text color="muted">Manage your workspace settings.</Text>
      <Divider className="my-5" />
      <Text className="mb-2" fontSize="lg" fontWeight="medium">
        Logo
      </Text>
      <Avatar
        className="mb-5 h-20"
        name="Joseph Mukorivo"
        rounded="lg"
        src="/complexus.png"
      />
      <Text className="mb-5" color="muted">
        Choose a logo for your workspace.
      </Text>

      <Input className="mb-4" label="Workspace name" required value="Joseph" />
      <Input
        className="mb-4"
        label="Workspace url"
        required
        value="https://test.complexus.app"
      />
      <Input className="mb-4" label="Company size" required value="Joseph" />
      <Button rounded="sm">Update workspace</Button>
    </Box>
  );
}
