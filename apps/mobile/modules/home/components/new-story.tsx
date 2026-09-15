import { useRouter } from "expo-router";
import { HeaderActions } from "@/components/ui/header-actions";
import { useTerminology } from "@/hooks/use-terminology";

export const NewStoryButton = () => {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  return (
    <HeaderActions
      createLabel={`Create ${getTermDisplay("storyTerm")}`}
      onCreate={() => router.push("/new")}
    />
  );
};
