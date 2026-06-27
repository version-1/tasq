"use client";

import { useTranslation } from "react-i18next";
import { DefaultLayout } from "@/components/layout/default";
import { PanelMessage } from "@/components/ui/pannel-message";

export function NotFoundRoute() {
  const { t } = useTranslation();

  return (
    <DefaultLayout>
      <PanelMessage
        title={t("notFound.title")}
        detail={t("notFound.detail")}
      />
    </DefaultLayout>
  );
}
