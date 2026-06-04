import { useTranslation } from "react-i18next";
import type { SupportedLanguage } from "@/lib/i18n";
import styles from "./index.module.css";

export function SettingsView({
  refreshIntervalMs,
  language,
  generatedAt,
  onRefreshIntervalChange,
  onLanguageChange,
}: {
  refreshIntervalMs: number;
  language: SupportedLanguage;
  generatedAt: string;
  onRefreshIntervalChange: (intervalMs: number) => void;
  onLanguageChange: (language: SupportedLanguage) => void;
}) {
  const { t } = useTranslation();

  return (
    <section className={styles.panelGrid}>
      <form className={`${styles.widePanel} ${styles.settingsForm}`} onSubmit={(event) => event.preventDefault()}>
        <h2>{t("settings.title")}</h2>
        <label>
          {t("settings.refreshInterval")}
          <select
            value={refreshIntervalMs}
            onChange={(event) => onRefreshIntervalChange(Number(event.target.value))}
          >
            <option value={1000}>{t("settings.refreshOneSecond")}</option>
            <option value={3000}>{t("settings.refreshThreeSeconds")}</option>
            <option value={5000}>{t("settings.refreshFiveSeconds")}</option>
            <option value={10000}>{t("settings.refreshTenSeconds")}</option>
          </select>
        </label>
        <label>
          {t("settings.language")}
          <select
            value={language}
            onChange={(event) => onLanguageChange(event.target.value as SupportedLanguage)}
          >
            <option value="ja">{t("settings.languageJapanese")}</option>
            <option value="en">{t("settings.languageEnglish")}</option>
          </select>
        </label>
        <label>
          {t("settings.apiOrigin")}
          <input value={apiOriginLabel()} readOnly />
        </label>
        <label>
          {t("settings.lastSummary")}
          <input value={generatedAt || t("common.pending")} readOnly />
        </label>
      </form>
    </section>
  );
}

function apiOriginLabel(): string {
  return `${window.location.origin}/api/tracker`;
}
