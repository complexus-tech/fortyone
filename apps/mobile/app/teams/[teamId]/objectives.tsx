import { Redirect, useLocalSearchParams } from "expo-router";
export default function ObjectivesScreen() {
  const { teamId } = useLocalSearchParams<{ teamId: string }>();
  return (
    <Redirect
      href={{
        pathname: "/teams/[teamId]",
        params: { teamId, section: "objectives" },
      }}
    />
  );
}
