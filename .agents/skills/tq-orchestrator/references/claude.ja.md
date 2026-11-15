# Claude Code の設定

Claude Code 内で、共通のサービス状態確認を通過した後に使う。コメントの author は `claude-code`。

- Monitor で選択済み CLI の watch を実行し、NDJSON の通知を受け取る。監視IDを保持し、`default.ja.md` に従ってイベントを扱う。
- Monitor がなければ不足する機能を報告し、管理不能な監視を始めない。Codex 固有のツール名を Claude Code に適用しない。
- 作業とレビューには Claude Code のサブエージェント機能を使う。停止時は自分の Monitor の処理を止めて終了を確認する。別途中止を指示されていなければ、委譲先と成果物を保持する。
