import { Link, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import styles from "./index.module.css";
import { breadcrumbSegmentsFromPathname, type BreadcrumbSegment } from "./segments";

export function Breadcrumb({ segments }: { segments?: BreadcrumbSegment[] } = {}) {
  const { t } = useTranslation();
  const { pathname } = useLocation();
  const items = segments ?? breadcrumbSegmentsFromPathname(pathname);

  return (
    <nav className={styles.breadcrumb} aria-label={t("header.breadcrumb")}>
      <ol className={styles.list}>
        {items.map((segment, index) => {
          const isLast = index === items.length - 1;

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
