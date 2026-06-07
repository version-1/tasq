import { useTranslation } from "react-i18next";
import styles from "./index.module.css";

type BreadcrumbProps = {
  projectName: string;
};

export function Breadcrumb({ projectName }: BreadcrumbProps) {
  const { t } = useTranslation();

  return (
    <div className={styles.breadcrumb} aria-label={t("header.breadcrumb")}>
      <span>{t("header.project")}</span>
      <span aria-hidden="true">/</span>
      <button className={styles.projectSwitch} type="button">
        {projectName}
        <span aria-hidden="true">⌄</span>
      </button>
    </div>
  );
}
