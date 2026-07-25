import { useTranslation } from "react-i18next";
import type { CommentType } from "@/lib/types";
import styles from "./index.module.css";

export const commentTypes = ["progress", "blocker", "handoff", "general"] satisfies CommentType[];

export function CommentTypeBadge({ type }: { type: CommentType }) {
  const { t } = useTranslation();

  return <span className={[styles.badge, styles[type]].join(" ")}>{t(`comments.types.${type}`)}</span>;
}
