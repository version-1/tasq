import { useTranslation } from "react-i18next";
import styles from "./index.module.css";

type HeaderButtonProps = {
  onAddTask: () => void;
};

export function HeaderButton({ onAddTask }: HeaderButtonProps) {
  const { t } = useTranslation();

  return (
    <>
      <button className={styles.primaryButton} type="button" onClick={onAddTask}>
        <span aria-hidden="true">＋</span>
        {t("header.addTask")}
      </button>
      <button className={styles.splitButton} type="button" aria-label={t("header.taskOptions")}>
        ⌄
      </button>
    </>
  );
}
