// nativewind.d.ts
declare module "nativewind" {
  export function useColorScheme(): {
    colorScheme: "light" | "dark";
    setColorScheme: (scheme: "light" | "dark" | "system") => void;
    toggleColorScheme: () => void;
  };
}
