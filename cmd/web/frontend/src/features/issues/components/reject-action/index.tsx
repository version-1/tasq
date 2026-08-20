import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  ContextMenu,
  ContextMenuGroupLabel,
  ContextMenuItem,
} from "@/components/ui/context-menu";
import { Button } from "@/components/ui/button";
import { IconProxy } from "@/components/ui/icon-proxy";
import {
  builtInChangeRequestShortcuts,
  type ChangeRequestShortcut,
} from "@/features/issues/change-request-shortcuts";
import styles from "./index.module.css";

export function RejectAction({
  disabled = false,
  onOpenDialog,
  onSelectShortcut,
  shortcuts = builtInChangeRequestShortcuts.reject,
}: {
  disabled?: boolean;
  onOpenDialog: () => void;
  onSelectShortcut: (shortcut: ChangeRequestShortcut) => Promise<void>;
  shortcuts?: readonly ChangeRequestShortcut[];
}) {
  const { t } = useTranslation();
  const menuID = useId();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isDisabled = disabled || isSubmitting;

  async function handleShortcut(shortcut: ChangeRequestShortcut) {
    setIsMenuOpen(false);
    setIsSubmitting(true);
    try {
      await onSelectShortcut(shortcut);
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className={styles.splitButton}>
      <Button
        className={styles.primaryButton}
        disabled={isDisabled}
        size="compact"
        variant="secondary"
        onClick={onOpenDialog}
      >
        {t("issues.reject.action")}
      </Button>
      <ContextMenu
        id={menuID}
        isOpen={isMenuOpen}
        label={t("issues.reject.shortcutMenu")}
        onOpenChange={setIsMenuOpen}
        trigger={(triggerProps) => (
          <button
            {...triggerProps}
            aria-label={t("issues.reject.shortcutMenu")}
            className={styles.menuButton}
            disabled={isDisabled}
            type="button"
          >
            <IconProxy name="chevron-down" size={14} />
          </button>
        )}
      >
        <ContextMenuGroupLabel>{t("issues.reject.action")}</ContextMenuGroupLabel>
        <ContextMenuItem
          disabled={isDisabled}
          onSelect={() => {
            setIsMenuOpen(false);
            onOpenDialog();
          }}
        >
          {t("issues.changeRequest.writeComment")}
        </ContextMenuItem>
        {shortcuts.map((shortcut) => (
          <ContextMenuItem
            key={shortcut.id}
            disabled={isDisabled}
            onSelect={() => void handleShortcut(shortcut)}
          >
            {shortcut.label}
          </ContextMenuItem>
        ))}
      </ContextMenu>
    </div>
  );
}
