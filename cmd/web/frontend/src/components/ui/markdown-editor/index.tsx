import { useEffect, useId, useState, type CSSProperties } from "react";
import { IconProxy } from "@/components/ui/icon-proxy";
import { Markdown } from "@/components/ui/markdown";
import styles from "./index.module.css";

type MarkdownEditorMode = "read" | "edit";
type MarkdownEditorTab = "raw" | "preview";

type MarkdownEditorLabels = {
  cancel: string;
  edit: string;
  empty: string;
  raw: string;
  preview: string;
  save: string;
  saving: string;
  textarea: string;
};

export function MarkdownEditor({
  error = "",
  initialMode = "read",
  initialTab = "raw",
  isSaving = false,
  labels,
  rows = 12,
  showActions = true,
  stablePanelRows,
  title,
  titleID,
  value,
  onChange,
  onSave,
}: {
  error?: string;
  initialMode?: MarkdownEditorMode;
  initialTab?: MarkdownEditorTab;
  isSaving?: boolean;
  labels: MarkdownEditorLabels;
  rows?: number;
  showActions?: boolean;
  stablePanelRows?: number;
  title?: string;
  titleID?: string;
  value: string;
  onChange?: (value: string) => void;
  onSave?: (value: string) => Promise<void> | void;
}) {
  const editorID = useId();
  const [mode, setMode] = useState<MarkdownEditorMode>(initialMode);
  const [activeTab, setActiveTab] = useState<MarkdownEditorTab>(initialTab);
  const [draft, setDraft] = useState(value);
  const [saveError, setSaveError] = useState("");
  const isEditing = mode === "edit";
  const errorMessage = saveError || error;
  const activePanelID = `${editorID}-panel`;
  const rawTabID = `${editorID}-raw-tab`;
  const previewTabID = `${editorID}-preview-tab`;
  const panelStyle = stablePanelRows
    ? ({
        "--markdown-editor-panel-min-block-size": `calc(${stablePanelRows}lh + (var(--space-3-5) * 2) + 2px)`,
      } as CSSProperties)
    : undefined;

  useEffect(() => {
    if (!isEditing) {
      setDraft(value);
    }
  }, [isEditing, value]);

  function handleEdit() {
    setDraft(value);
    setSaveError("");
    setActiveTab(initialTab);
    setMode("edit");
  }

  function handleCancel() {
    setDraft(value);
    setSaveError("");
    setActiveTab(initialTab);
    setMode("read");
  }

  function handleDraftChange(nextValue: string) {
    setSaveError("");
    setDraft(nextValue);
    onChange?.(nextValue);
  }

  async function handleSave() {
    if (!onSave) {
      return;
    }
    try {
      setSaveError("");
      await onSave(draft);
      setMode("read");
    } catch (caughtError) {
      setSaveError(caughtError instanceof Error ? caughtError.message : String(caughtError));
    }
  }

  return (
    <div className={styles.editor}>
      {title ? (
        <div className={styles.header}>
          <h3 id={titleID}>{title}</h3>
          {!isEditing && showActions ? (
            <button
              type="button"
              className={styles.iconButton}
              aria-label={labels.edit}
              onClick={handleEdit}
            >
              <IconProxy name="pencil" size={16} />
            </button>
          ) : null}
        </div>
      ) : null}

      {isEditing ? (
        <div className={styles.editSurface}>
          <div className={styles.editHeader}>
            <div className={styles.tabs} role="tablist">
              <button
                type="button"
                className={activeTab === "raw" ? styles.activeTab : styles.tab}
                aria-controls={activePanelID}
                aria-selected={activeTab === "raw"}
                id={rawTabID}
                role="tab"
                tabIndex={activeTab === "raw" ? 0 : -1}
                onClick={() => setActiveTab("raw")}
              >
                {labels.raw}
              </button>
              <button
                type="button"
                className={activeTab === "preview" ? styles.activeTab : styles.tab}
                aria-controls={activePanelID}
                aria-selected={activeTab === "preview"}
                id={previewTabID}
                role="tab"
                tabIndex={activeTab === "preview" ? 0 : -1}
                onClick={() => setActiveTab("preview")}
              >
                {labels.preview}
              </button>
            </div>
            {showActions ? (
              <div className={styles.actions}>
                <button type="button" onClick={handleCancel} disabled={isSaving}>
                  {labels.cancel}
                </button>
                <button
                  type="button"
                  className={styles.saveButton}
                  onClick={() => void handleSave()}
                  disabled={isSaving}
                >
                  {isSaving ? labels.saving : labels.save}
                </button>
              </div>
            ) : null}
          </div>
          {activeTab === "raw" ? (
            <div
              className={styles.panel}
              id={activePanelID}
              role="tabpanel"
              aria-labelledby={rawTabID}
              style={panelStyle}
            >
              <textarea
                aria-label={labels.textarea}
                className={styles.textarea}
                rows={rows}
                value={draft}
                onChange={(event) => handleDraftChange(event.target.value)}
              />
            </div>
          ) : (
            <div
              className={`${styles.panel} ${styles.preview}`}
              id={activePanelID}
              role="tabpanel"
              aria-labelledby={previewTabID}
              style={panelStyle}
            >
              <Markdown content={draft} emptyText={labels.empty} />
            </div>
          )}
          {errorMessage ? <p className={styles.error}>{errorMessage}</p> : null}
        </div>
      ) : (
        <Markdown content={value} emptyText={labels.empty} />
      )}
    </div>
  );
}
