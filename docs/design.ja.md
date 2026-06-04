# Tasq 設計

Tasq design documentation は concern ごとに分割しています。

- [Architecture](design/architecture.ja.md): ownership boundary、component、dependency、state ownership、current MVP behavior。
- [API](design/api.ja.md): issue-tracker API surface と CLI command mapping。
- [Configuration](design/configuration.ja.md): local home directory、runtime state、Compose development configuration。
- [Schema](design/schema.md): entity field specifications と validation rules。
- [Operations](design/operations.ja.md): local development environment、verification、open decision。
- [Deployment](design/deployment.ja.md): release tag 作成、GitHub Actions、GoReleaser の deployment flow。

Web UI の構造と styling convention は [design/web.md](design/web.md) を参照してください。
