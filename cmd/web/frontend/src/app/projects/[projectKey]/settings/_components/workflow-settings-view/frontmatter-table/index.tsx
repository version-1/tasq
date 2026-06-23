import type { CSSProperties } from "react";
import { useTranslation } from "react-i18next";
import type { FrontmatterRow } from "./rows";
import styles from "./index.module.css";

type FrontmatterTableProps = {
  rows: FrontmatterRow[];
};

export function FrontmatterTable({ rows }: FrontmatterTableProps) {
  const { t } = useTranslation();

  return (
    <div className={styles.tableWrap}>
      <table className={styles.frontmatterTable}>
        <thead>
          <tr>
            <th className={styles.tableHeadCell} scope="col">
              {t("projectSettings.frontmatterKey")}
            </th>
            <th className={styles.tableHeadCell} scope="col">
              {t("projectSettings.frontmatterValue")}
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => (
            <tr key={row.id}>
              <th className={styles.keyCell} scope="row">
                <span
                  className={`${styles.keyContent} ${
                    row.kind === "branch" ? styles.branchKey : ""
                  }`}
                  style={{ "--frontmatter-depth": row.depth } as CSSProperties}
                >
                  <span className={styles.keyText}>{row.key}</span>
                </span>
              </th>
              <td className={styles.valueCell}>{row.value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
