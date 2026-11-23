import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { Breadcrumb } from "@/components/layout/header/breadcrumb";
import { Button } from "@/components/ui/button";
import { Markdown } from "@/components/ui/markdown";
import { MarkdownEditor } from "@/components/ui/markdown-editor";
import { fetchProjectDocumentArtifact, updateProjectDocumentArtifact } from "@/lib/api";
import type { DocumentArtifact, UpdateDocumentArtifactInput } from "@/lib/types";
import styles from "./index.module.css";

type DocumentDetailViewProps = {
  projectID?: number;
  projectKey: string;
  documentID: string;
  projectError?: string;
};

export function DocumentDetailView({
  projectID,
  projectKey,
  documentID,
  projectError,
}: DocumentDetailViewProps) {
  const { t } = useTranslation();
  const [document, setDocument] = useState<DocumentArtifact | null>(null);
  const [loadError, setLoadError] = useState("");
  const [draft, setDraft] = useState<{ title: string; body: string } | null>(null);
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (projectID === undefined) return;
    let active = true;
    async function load(id: number) {
      try {
        const result = await fetchProjectDocumentArtifact(id, documentID, { silent: true });
        if (active) setDocument(result);
      } catch (error) {
        if (active) setLoadError(error instanceof Error ? error.message : t("projectArtifacts.failedToLoad"));
      }
    }
    void load(projectID);
    return () => { active = false; };
  }, [projectID, documentID, t]);

  function startEditing(item: DocumentArtifact) {
    setSaved(false);
    setSaveError("");
    setDraft({ title: item.title, body: item.body ?? "" });
  }

  const titleChanged = draft !== null && draft.title.trim() !== document?.title;
  const bodyChanged = draft !== null && draft.body !== document?.body;
  const canSave = draft !== null && Boolean(draft.title.trim() && draft.body)
    && (titleChanged || bodyChanged) && !saving;

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!draft || projectID === undefined || !canSave) return;
    const changes: UpdateDocumentArtifactInput = {};
    if (titleChanged) changes.title = draft.title;
    if (bodyChanged) changes.body = draft.body;
    setSaving(true);
    setSaveError("");
    try {
      const updated = await updateProjectDocumentArtifact(projectID, documentID, changes, { silent: true });
      setDocument(updated);
      setDraft(null);
      setSaved(true);
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : t("projectArtifacts.failedToSave"));
    } finally {
      setSaving(false);
    }
  }

  const error = projectError || loadError;
  return (
    <main className={styles.page}>
      <Breadcrumb segments={[
        { label: t("header.artifacts"), href: `/projects/${encodeURIComponent(projectKey)}/artifacts` },
        { label: document?.title ?? t("projectArtifacts.document") },
      ]} />
      {error ? (
        <p className={styles.error} role="alert">{error}</p>
      ) : !document ? (
        <p role="status">{t("projectArtifacts.loading")}</p>
      ) : (
        <article className={styles.document}>
          <div className={styles.heading}>
            <h1>{document.title}</h1>
            {!draft && (
              <Button variant="secondary" onClick={() => startEditing(document)}>
                {t("projectArtifacts.edit")}
              </Button>
            )}
          </div>
          {saved && <p role="status">{t("projectArtifacts.saved")}</p>}
          {draft ? (
            <form className={styles.form} onSubmit={(event) => void save(event)}>
              <label className={styles.field}>
                {t("projectArtifacts.documentTitle")}
                <input
                  autoFocus
                  required
                  readOnly={saving}
                  value={draft.title}
                  onChange={(event) => setDraft({ ...draft, title: event.target.value })}
                />
              </label>
              <MarkdownEditor
                initialMode="edit"
                showActions={false}
                readOnly={saving}
                value={draft.body}
                onChange={(body) => setDraft({ ...draft, body })}
                labels={{
                  cancel: t("markdownEditor.cancel"),
                  edit: t("markdownEditor.edit"),
                  empty: "",
                  preview: t("markdownEditor.preview"),
                  raw: t("markdownEditor.raw"),
                  save: t("markdownEditor.save"),
                  saving: t("markdownEditor.saving"),
                  textarea: t("projectArtifacts.documentBody"),
                }}
              />
              {saveError && <p className={styles.error} role="alert">{saveError}</p>}
              <div className={styles.actions}>
                <Button variant="secondary" disabled={saving} onClick={() => setDraft(null)}>
                  {t("markdownEditor.cancel")}
                </Button>
                <Button type="submit" disabled={!canSave}>
                  {saving ? t("markdownEditor.saving") : t("markdownEditor.save")}
                </Button>
              </div>
            </form>
          ) : <Markdown content={document.body ?? ""} emptyText="" />}
        </article>
      )}
    </main>
  );
}
