import { useTranslation } from "react-i18next";
import type { Artifact } from "@/lib/types";
import { pullRequestArtifact } from "@/features/issues/artifacts";
import styles from "./index.module.css";

export function ArtifactsSection({ artifacts }: { artifacts: readonly Artifact[] }) {
  const { t } = useTranslation();
  const pullRequest = pullRequestArtifact(artifacts);

  if (!pullRequest) {
    return null;
  }

  return (
    <section className={styles.section} aria-labelledby="artifacts-heading">
      <h2 id="artifacts-heading" className={styles.heading}>{t("issues.detailPage.artifacts")}</h2>
      <a
        className={styles.link}
        href={pullRequest.data_value}
        rel="noopener noreferrer"
        target="_blank"
      >
        {t("issues.detailPage.pullRequest")}
      </a>
    </section>
  );
}
