import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { fetchProjectDocumentArtifacts } from "@/lib/api";
import type { DocumentArtifact } from "@/lib/types";
import styles from "./index.module.css";

type DocumentArtifactsViewProps = {
  projectID: number;
  projectKey: string;
};

export function DocumentArtifactsView({ projectID, projectKey }: DocumentArtifactsViewProps) {
  const { t } = useTranslation();
  const [documents, setDocuments] = useState<DocumentArtifact[]>([]);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError("");

    const load = async () => {
      try {
        const items = await fetchProjectDocumentArtifacts(projectID, { silent: true });
        if (!active) return;
        setDocuments(items);
      } catch (reason: unknown) {
        if (active) setError(reason instanceof Error ? reason.message : t("projectArtifacts.failedToLoad"));
      } finally {
        if (active) setLoading(false);
      }
    };
    void load();
    return () => { active = false; };
  }, [projectID, t]);

  if (loading) return <p className={styles.message} role="status">{t("projectArtifacts.loading")}</p>;
  if (error) return <p className={styles.error} role="status">{error}</p>;

  return (
    <section className={styles.layout} aria-labelledby="project-artifacts-heading">
      <div className={styles.headingRow}>
        <h1 id="project-artifacts-heading">{t("projectArtifacts.title")}</h1>
        <span>{documents.length}</span>
      </div>
      {documents.length === 0 ? <p className={styles.message}>{t("projectArtifacts.empty")}</p> : (
        <nav aria-label={t("projectArtifacts.navigation")} className={styles.list}>
          {documents.map((document) => <Link className={styles.link} key={document.id} to={`/projects/${encodeURIComponent(projectKey)}/artifacts/${encodeURIComponent(document.id)}`}>
            <span>{document.title}</span>
            <small>{document.type === "work_report" ? t("issues.detailPage.workReport") : t("issues.detailPage.referenceDocument")}</small>
          </Link>)}
        </nav>
      )}
    </section>
  );
}
