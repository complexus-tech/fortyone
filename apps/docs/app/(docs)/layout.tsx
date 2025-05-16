// import { DocsLayout } from "fumadocs-ui/layouts/docs";
// import type { ReactNode } from "react";
// import { baseOptions } from "@/app/layout.config";
// import { source } from "@/lib/source";

// export default function Layout({ children }: { children: ReactNode }) {
//   return (
//     <DocsLayout tree={source.pageTree} {...baseOptions}>
//       {children}
//     </DocsLayout>
//   );
// }

import { DocsLayout } from "fumadocs-ui/layouts/notebook";
import { baseOptions } from "@/app/layout.config";
import { source } from "@/lib/source";
import type { ReactNode } from "react";
import { Library } from "lucide-react";
export default function Layout({ children }: { children: ReactNode }) {
  return (
    <DocsLayout
      {...baseOptions}
      // the position of navbar
      nav={{ ...baseOptions.nav, mode: "top" }}
      // the position of Sidebar Tabs
      tabMode="navbar"
      themeSwitch={{
        mode: "light-dark-system",
      }}
      links={[
        {
          text: "Sign up",
          url: "https://www.complexus.app/signup",
        },
        {
          text: "Login",
          url: "https://www.complexus.app/login",
        },
      ]}
      tree={source.pageTree}
    >
      {children}
    </DocsLayout>
  );
}
