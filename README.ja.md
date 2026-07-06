# tasq

AI coding agent task manager.

tasq は、実装作業を見える queue にし、その queue のための local services を起動し、`tq` CLI と Web UI の両方から進捗を確認できるようにする tool です。

English counterpart: [README.md](README.md).

## Problem

AI coding agents により、複数の実装タスクを同時に進められるようになりました。ボトルネックは code generation そのものから、parallel work の coordination へ移ります。

チームは、どのタスクが存在するか、どれが ready か、何が実行中か、何を review すべきか、どの local workspace がどの task に対応するかを把握し続ける必要があります。すべての agent を 1 つの checkout で実行すると、branch switching や file conflict のリスクも高まります。

## Solution

tasq は agent work に product surface を与えます。Issue tracker、local services、CLI、Web UI により、task state と project context を 1 か所に集めます。

![Tasq task queue to parallel agent workspaces](docs/site/static/img/agent-task-queue.svg)

タスクは review 可能な workflow を進みます。

```text
backlog -> ready -> in_progress -> review -> done
```

## Features

- Agent-sized tasks、priorities、dependencies、comments を扱う issue queue。
- Task 作成、状態更新、progress comment 追加、workflow scripting のための `tq` CLI。
- SQLite backed の local issue-tracker、orchestrator、Web UI services。
- Projects、issues、comments、queue status、service state を確認する Web UI。
- Issues を実際の local repository path に結び付ける project registration。
- Binary-only local setup に必要な runtime binaries を含む release archives。

## Install

最新の GitHub Release archive を platform に合わせて download し、展開した 4 つの binaries を `PATH` に配置します。

各 release tarball には次の binary が含まれます。

- `tq`: 直接実行する CLI。
- `issue-tracker`: local REST API service。
- `orchestrator`: local run-state service。
- `web`: frontend assets を埋め込んだ local Web UI server。

最新の formal release を install します。

```sh
version="$(curl -fsSLI -o /dev/null -w '%{url_effective}' https://github.com/version-1/tasq/releases/latest | sed 's#.*/##')"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in x86_64) arch="amd64" ;; aarch64) arch="arm64" ;; esac
archive="tasq_${version#v}_${os}_${arch}.tar.gz"
tmp_dir="$(mktemp -d)"
curl -fsSL "https://github.com/version-1/tasq/releases/download/${version}/${archive}" -o "${tmp_dir}/${archive}"
tar -xzf "${tmp_dir}/${archive}" -C "${tmp_dir}"
install_dir="${HOME}/.local/bin"
mkdir -p "$install_dir"
cp "${tmp_dir}/tq" "${tmp_dir}/issue-tracker" "${tmp_dir}/orchestrator" "${tmp_dir}/web" "$install_dir/"
chmod 0755 "$install_dir/tq" "$install_dir/issue-tracker" "$install_dir/orchestrator" "$install_dir/web"
```

Install directory が `PATH` に含まれることを確認します。

```sh
export PATH="${HOME}/.local/bin:${PATH}"
tq version
```

## Getting Started

Tasq は machine-local runtime data を `TQ_HOME` 配下に保存します。`TQ_HOME` が未設定の場合は `~/.tasq` を使います。

Local databases を初期化し、明示的な local ports で issue-tracker、orchestrator、Web UI を起動します。

```sh
export TQ_HOME="${HOME}/.tasq"
export TQ_API_URL="http://127.0.0.1:47651"
tq migrate
issue-tracker -addr 127.0.0.1:47651 &
tracker_pid=$!
orchestrator -issue-tracker http://127.0.0.1:47651 -port 47652 &
orchestrator_pid=$!
web -addr 127.0.0.1:47653 \
  -tracker-url http://127.0.0.1:47651 \
  -orchestrator-url http://127.0.0.1:47652 &
web_pid=$!
sleep 1
```

この手順では local services を次の ports で起動します。

| Service | Port |
| --- | ---: |
| issue-tracker | `47651` |
| orchestrator | `47652` |
| web | `47653` |

Local repository を project として登録します。

```sh
tq project add --key tasq-demo .
```

Task を作成し、queue に載せます。

```sh
tq issue create \
  --project tasq-demo \
  --title "Try Tasq from binaries" \
  --description "Create the first issue through the tq CLI."
tq issue list --project tasq-demo
```

Web UI を [http://127.0.0.1:47653](http://127.0.0.1:47653) で開き、project と issue が表示されることを確認します。

完了したら local services を停止します。

```sh
kill "$web_pid" "$orchestrator_pid" "$tracker_pid"
```

これらの ports のいずれかが使用中の場合は、別の loopback port を選び、`TQ_API_URL`、`-issue-tracker`、`-tracker-url`、`-orchestrator-url` を揃えて変更してください。

## Documentation

- [Design documentation](docs/design.ja.md)
- [Release binary startup notes](docs/design/release-binary-startup.ja.md)
- [CLI reference](docs/site/i18n/ja/docusaurus-plugin-content-docs/current/reference/cli-reference.md)

## Development

Repository workflow、local development、verification は [docs/development.ja.md](docs/development.ja.md) を参照してください。
