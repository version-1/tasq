import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import styles from "./index.module.css";
import { breadcrumbSegmentsFromPathname } from "./segments";

export function Breadcrumb() {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const segments = breadcrumbSegmentsFromPathname(pathname);

  return (
    <nav className={styles.breadcrumb} aria-label={t("header.breadcrumb")}>
      <ol className={styles.list}>
        {segments.map((segment, index) => {
          const isLast = index === segments.length - 1;

          return (
            <li className={styles.item} key={`${segment.label}-${index}`}>
              {index > 0 ? (
                <span className={styles.separator} aria-hidden="true">
                  &rsaquo;
                </span>
              ) : null}
              {segment.href && !isLast ? (
                <Link className={styles.link} to={segment.href}>
                  {segment.label}
                </Link>
              ) : (
                <span className={styles.current} aria-current={isLast ? "page" : undefined}>
                  {segment.label}
                </span>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
