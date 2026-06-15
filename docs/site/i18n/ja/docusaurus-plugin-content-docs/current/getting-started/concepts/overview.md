---
id: overview
title: 概要
sidebar_position: 1
---

# 概念の概要

Tasq は issue state と run state を分離します。issue-tracker は users と agents が取り組んでいるものを所有します。orchestrator は agent runs、workspaces、runtime events の記録方法を所有します。

![Tasq の概念概要](/img/concepts-overview-ja.svg)

## 所有モデル

issue-tracker は projects、issues、comments、attachments、board summaries の user-facing source of truth です。`tq` や Web UI などの clients は issue-tracker API を通じてその state を読み書きします。terminal client は計画されていますが、まだ published user guide には含まれていません。

orchestrator は runs、runner events、workspace metadata の runtime source of truth です。issue status を直接変更しません。task status が変わる場合でも、その変更は issue-tracker を通ります。

## クライアントの流れ

```mermaid
flowchart LR
  CLI[tq CLI] --> Tracker[Issue Tracker API]
  Web[Web UI] --> Tracker
  Terminal[Terminal client planned] -. planned .-> Tracker
  Tracker --> IssueDB[(issues.sqlite)]
  Tracker --> Attachments[$TQ_HOME attachments]
  Orchestrator[Orchestrator API] --> RunDB[(orchestrator.sqlite)]
  Orchestrator --> Workspaces[Issue workspaces]
  Web -. runtime views .-> Orchestrator
```

## 状態の境界

Issue status と run status は意図的に異なる概念です。

| Area | Owner | Examples |
| --- | --- | --- |
| Issue workflow | Issue Tracker | `backlog`, `ready`, `in_progress`, `review`, `done` |
| Run lifecycle | Orchestrator | `queued`, `running`, `waiting_for_input`, `succeeded`, `failed` |
| Workspace metadata | Orchestrator | workspace path, setup result, source path |
| Attachments | Issue Tracker | `TQ_HOME` 配下の image metadata と bytes |

この分離により、orchestration internals が進化しても user workflow は安定したままになります。
