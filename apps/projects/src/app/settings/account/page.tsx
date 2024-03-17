import { Avatar, Box, Button, Divider, Input, Text } from "ui";

export default function Page(): JSX.Element {
  return (
    <Box>
      <Text className="mb-3" fontSize="3xl">
        Profile
      </Text>
      <Text color="muted">Manage your profile.</Text>
      <Divider className="my-5" />
      <Text className="mb-2" color="muted" fontSize="lg">
        Your profile icon
      </Text>
      <Avatar
        className="mb-5 h-28"
        name="Joseph Mukorivo"
        src="https://images.unsplash.com/photo-1677576874778-df95ea6ff733?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHx0b3BpYy1mZWVkfDI4fHRvd0paRnNrcEdnfHxlbnwwfHx8fHw%3D&auto=format&fit=crop&w=800&q=60"
      />

      <Input className="mb-4" label="First name" required value="Joseph" />
      <Input className="mb-4" label="Last name" required value="Mukorivo" />
      <Input
        className="mb-4"
        label="Email"
        type="email"
        value="joseph@complexus.tech"
      />
      <Input className="mb-4" label="Role" value="Developer" />
      <Input className="mb-4" label="Username" value="josemukorivo" />
      <Input className="mb-6" label="Timezone" value="Nairobi / GMT + 3" />
      <Button rounded="sm">Update profile</Button>
    </Box>
  );
}
