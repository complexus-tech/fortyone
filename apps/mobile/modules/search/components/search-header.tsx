import { FilterIcon } from "@/components/icons/filter";
import { Keyboard, View } from "react-native";
import { useRouter } from "expo-router";
import { Back, HeaderActions, ScreenHeader } from "@/components/ui";
import { useTerminology } from "@/hooks/use-terminology";

type HeaderProps = {
  objectivesEnabled: boolean;
  searchType: "stories" | "objectives";
  setSearchType: (type: "stories" | "objectives") => void;
};

export function SearchHeader({
  objectivesEnabled,
  searchType,
  setSearchType,
}: HeaderProps) {
  const router = useRouter();
  const { getTermDisplay } = useTerminology();
  const scopes = [
    {
      value: "stories" as const,
      label: getTermDisplay("storyTerm", {
        variant: "plural",
        capitalize: true,
      }),
    },
    {
      value: "objectives" as const,
      label: getTermDisplay("objectiveTerm", {
        variant: "plural",
        capitalize: true,
      }),
    },
  ].filter((scope) => objectivesEnabled || scope.value === "stories");
  const selectedLabel = getTermDisplay(
    searchType === "stories" ? "storyTerm" : "objectiveTerm",
    { variant: "plural" },
  );

  return (
    <View>
      <ScreenHeader
        title="Search"
        compact
        leading={
          <Back
            onPress={() => {
              Keyboard.dismiss();
              if (router.canGoBack()) {
                router.back();
              } else {
                router.replace("/");
              }
            }}
          />
        }
        trailing={
          <HeaderActions
            createLabel={`Create ${getTermDisplay("storyTerm")}`}
            onCreate={() => router.push("/new")}
            menuLabel={`Search type: ${selectedLabel}`}
            menuContent={<FilterIcon size={22} />}
            actions={scopes.map((scope) => ({
              label: scope.label,
              selected: scope.value === searchType,
              onPress: () => setSearchType(scope.value),
            }))}
          />
        }
      />
    </View>
  );
}
