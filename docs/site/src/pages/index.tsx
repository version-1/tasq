import clsx from 'clsx';
import Heading from '@theme/Heading';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';

type HomeContent = {
  description: string;
  primaryAction: string;
  secondaryAction: string;
  features: Array<{
    title: string;
    description: string;
  }>;
};

const contentByLocale: Record<string, HomeContent> = {
  en: {
    description:
      'Tasq is a local-first issue tracker and workflow tool for teams that coordinate human work, CLI automation, and agent runs in one repository.',
    primaryAction: 'Get started',
    secondaryAction: 'Read CLI reference',
    features: [
      {
        title: 'Local-first issue state',
        description:
          'Track issues, comments, projects, and attachments in a host-local SQLite-backed service.',
      },
      {
        title: 'Agent-friendly CLI',
        description:
          'Use tq from scripts, agents, and terminals with text output for humans and JSON output for tools.',
      },
      {
        title: 'Workflow-aware operations',
        description:
          'Run local services, inspect logs, validate project workflow files, and keep orchestration state separate from issue state.',
      },
    ],
  },
  ja: {
    description:
      'Tasq は、人の作業、CLI 自動化、エージェント実行を 1 つのリポジトリで扱うための local-first な issue tracker と workflow tool です。',
    primaryAction: 'はじめる',
    secondaryAction: 'CLI Reference を読む',
    features: [
      {
        title: 'ローカル優先の issue 状態',
        description:
          'issue、comment、project、attachment をホストローカルの SQLite backed service で管理します。',
      },
      {
        title: 'エージェントが扱いやすい CLI',
        description:
          'tq は script、agent、terminal から使え、human-readable output と tool 向け JSON output を提供します。',
      },
      {
        title: 'Workflow を意識した運用',
        description:
          'local service の起動、log 確認、project workflow file の検証を行い、orchestration state と issue state を分離します。',
      },
    ],
  },
};

export default function Home(): JSX.Element {
  const {i18n, siteConfig} = useDocusaurusContext();
  const content = contentByLocale[i18n.currentLocale] ?? contentByLocale.en;

  return (
    <Layout
      title={`${siteConfig.title} Documentation`}
      description={content.description}>
      <main>
        <section className={styles.hero}>
          <div className={styles.heroInner}>
            <Heading as="h1" className={styles.heroTitle}>
              Tasq
            </Heading>
            <p className={styles.heroDescription}>{content.description}</p>
            <div className={styles.heroActions}>
              <Link
                className={clsx('button button--primary button--lg', styles.heroButton)}
                to="/docs/getting-started">
                {content.primaryAction}
              </Link>
              <Link
                className={clsx('button button--secondary button--lg', styles.heroButton)}
                to="/docs/cli-reference">
                {content.secondaryAction}
              </Link>
            </div>
          </div>
        </section>
        <section className={styles.features}>
          <div className={styles.featureGrid}>
            {content.features.map((feature) => (
              <article className={styles.feature} key={feature.title}>
                <Heading as="h2" className={styles.featureTitle}>
                  {feature.title}
                </Heading>
                <p className={styles.featureDescription}>{feature.description}</p>
              </article>
            ))}
          </div>
        </section>
      </main>
    </Layout>
  );
}
