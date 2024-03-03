import { Divider, Text } from "ui";

export default function Page(): JSX.Element {
  return (
    <div>
      <Text className="mb-2" fontSize="3xl">
        Settings
      </Text>
      <Text color="muted">
        Lorem ipsum dolor sit amet consectetur, adipisicing elit. Aspernatur
        esse numquam atque perspiciatis repellat! Maiores.
      </Text>
      <Divider className="my-4" />
    </div>
  );
}
