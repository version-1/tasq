import { createContext, useContext, type ReactNode } from "react";
import { useParams } from "react-router-dom";

export type TabPageLink = {
  href: string;
  key: string;
  titleKey: string;
};

type TabsContextValue = {
  activeKey: string | null;
  pages: readonly TabPageLink[];
};

const emptyTabs: TabsContextValue = {
  activeKey: null,
  pages: [],
};

const tabsContext = createContext<TabsContextValue>(emptyTabs);

export function TabsProvider({
  activeKey,
  children,
  pages,
}: {
  activeKey?: string | null;
  children: ReactNode;
  pages?: readonly TabPageLink[];
}) {
  const params = useParams();
  const resolvedPages = (pages ?? []).map((page) => ({
    ...page,
    href: resolveTabHref(page.href, params),
  }));

  return (
    <tabsContext.Provider value={{ activeKey: activeKey ?? null, pages: resolvedPages }}>
      {children}
    </tabsContext.Provider>
  );
}

export function useTabs(): TabsContextValue {
  return useContext(tabsContext);
}

function resolveTabHref(
  href: string,
  params: Readonly<Record<string, string | undefined>>,
): string {
  return Object.entries(params).reduce((current, [key, value]) => {
    const replacement = value ?? "";
    return current
      .replaceAll(`:${key}`, replacement)
      .replaceAll(`{${key}}`, replacement);
  }, href);
}
