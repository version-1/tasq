---
name: fx-strategy-optimizer
description: Develop, register, and evaluate a new fx-autotrader strategy and its currency-pair profile against a requested profit-factor target. Use this skill whenever the user asks to create or improve a strategy/profile through backtests, tune entry conditions for PF, win rate, or drawdown, or requests a backtest performance report. Require a target PF, date range, minimum trade count, maximum-drawdown condition, and currency pair before changing code or running a backtest.
---

# FX Strategy Optimizer

Create a maintainable `fx-autotrader` strategy and profile, then iteratively evaluate it until the requested performance criteria are met or the available data/code makes further progress impossible. Treat a backtest as research, not evidence of future profitability.

## Required inputs

Collect all of these before implementation. Ask only for missing or ambiguous values.

| Input | Accept | Notes |
| --- | --- | --- |
| Target PF | Positive number | The threshold to meet on the holdout period. |
| Target period | Inclusive calendar dates or RFC3339 timestamps | Convert calendar dates to the repository's JST FX-market boundaries and state the exact timestamps used. |
| Minimum trade count | Positive integer | Apply it to the same scope as PF. |
| Maximum-drawdown condition | Absolute amount or percent plus its baseline | The stored metric is an absolute loss. A percent requires the configured backtest investment as its baseline. |
| Currency pair | Canonical pair key, such as `usdjpy` | Convert display forms such as `USD_JPY` or `USD/JPY` to the canonical key before using repository paths or the CLI. |

Also ask for a strategy/profile name if the user has a preference. Otherwise choose a concise, descriptive kebab-case profile name and the matching snake_case strategy name.

## Repository contract

Work in the `fx-autotrader` worktree. Read these before modifying code because they are the source of truth for the current contracts:

- `AGENTS.md`
- `docs/guide/strategy.md`
- `docs/guide/profile.md`
- `docs/guide/run-backtest.md`
- `internal/strategy/base.go`
- `internal/profile/loader.go`

Keep responsibilities separate:

- Put entry decisions, time filters, position filters, regime filters, and signal construction in `internal/strategy/...` using `strategy.BaseStrategy`.
- Keep order sizing, take-profit/stop-loss, and instrument constraints in `internal/profile/profiles/<currency-pair>/<profile-name>.json`.
- Register every new strategy name in `internal/profile/loader.go`; a JSON profile alone cannot make a new strategy executable.
- Do not alter live runtime settings, credentials, historical prices, or existing profiles to make a backtest look better.

## Workflow

### 1. Establish a reproducible baseline

1. Inspect the available historical-rate coverage for the requested pair and period using the runbook's query.
2. Inspect the closest existing strategy and profile. Record its profile name, code location, order parameters, and baseline results.
3. Split the requested period chronologically into an optimization segment and a final holdout segment. Default to 70% / 30%, preserving market-week boundaries when possible. If the requested period is too short to produce the minimum trades in both segments, explain that a trustworthy holdout evaluation is not possible and ask whether to widen the period or explicitly accept in-sample-only research.
4. Fix the investment, sizing method, and all non-strategy runtime settings for the comparison. State them in the report.

The holdout prevents repeatedly selecting conditions that merely fit the same past price path. Do not report the target as achieved unless the holdout satisfies every requested threshold.

### 2. Implement one focused candidate

1. Start from the closest existing strategy's conditions. Form a short hypothesis that explains which market behavior the change is intended to capture or reject.
2. Add a small strategy package that returns `*strategy.BaseStrategy` from `NewStrategy`. Compose conditions with the existing filter interfaces and use a `CommandBuilder` only when the reference-price behavior differs.
3. Register the strategy in `resolveStrategy` in `internal/profile/loader.go`.
4. Add a profile at `internal/profile/profiles/<currency-pair>/<profile-name>.json` with a precise description. Use the canonical currency-pair directory and a supported strategy name.
5. Add or extend tests for the strategy's acceptance/rejection behavior, warmup requirement where relevant, and loader coverage. Run the focused Go tests before a backtest.

Keep each iteration attributable: change one coherent hypothesis at a time. Avoid broad parameter sweeps that produce an unexplained winner.

### 3. Backtest and collect metrics

Determine the Compose project name from the target worktree before running a backtest. Use the worktree's isolated project name explicitly, rather than a fixed repository-wide value, so its network, containers, and volumes cannot collide with another workspace. The Makefile derives a default from the current directory name, but record the explicit value used in the report.

Run the repository's documented command with that project name and an isolated profile name:

```bash
COMPOSE_PROJECT_NAME=<worktree-compose-project-name> make backtest \
  FROM=<from-rfc3339> \
  TO=<to-rfc3339> \
  PROFILE=<profile-name> \
  CURRENCY_PAIR=<canonical-currency-pair>
```

Capture `result_id` from the completed run. Query `backtest_result_runs`, `backtest_result_slices`, and `backtest_result_trades` as documented in `docs/guide/run-backtest.md`. Collect at least:

- trade count and win rate
- profit factor, gross profit, gross loss, and expected value
- final profit and maximum drawdown
- recovery factor when defined
- date range, currency pair, profile name, strategy name, result ID, and code revision

Treat a missing PF (no losing trades) as undefined, not automatically as passing. Treat a zero-trade run, missing data, or failed run as a failed evaluation and diagnose it before changing conditions.

### 4. Decide and iterate

Evaluate the holdout in this order:

1. minimum trade count
2. maximum-drawdown condition
3. target PF

When a criterion fails, use the filter traces, trade outcomes, and baseline comparison to choose the next single hypothesis. Re-run focused tests, the optimization segment, and then the holdout for each candidate. Preserve the best candidate and its evidence; do not overwrite an existing profile.

Continue while the target remains unmet and a concrete, evidence-based next hypothesis exists. If progress stalls, data is insufficient, or the requested criteria conflict, stop and report the evidence and the least risky next experiment. Never claim success by silently relaxing a requested criterion.

## Completion checks

Before declaring success, verify all of the following:

- The strategy is constructed through `strategy.BaseStrategy` and registered by the profile loader.
- The new profile loads for the requested canonical currency pair.
- Focused tests and `go test ./...` pass, unless an unrelated pre-existing failure is clearly documented.
- The holdout has valid data and meets target PF, minimum trade count, and maximum-drawdown condition.
- No live-trading config, secret, or historical price data was changed.

## Required final report

Use this structure. Include failed iterations compactly so the result is reproducible.

```markdown
# <strategy/profile name> backtest report

## Decision
- Status: achieved | not achieved | blocked
- Requested criteria: PF >= <target>, trades >= <minimum>, max DD <condition>
- Holdout verdict: pass | fail

## Artifacts
- Strategy: <path>
- Profile: <path>
- Registration and tests: <paths>

## Reproducibility
- Currency pair: <canonical key>
- Optimization period: <from> to <to>
- Holdout period: <from> to <to>
- Command: `<exact command>`
- Compose project name: `<worktree-compose-project-name>`
- Result IDs: <optimization and holdout IDs>
- Code revision: <commit or dirty-worktree note>

## Metrics
| Scope | Trades | Win rate | PF | Max DD | Final profit | Expected value | Verdict |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Baseline holdout | | | | | | | |
| Candidate holdout | | | | | | | |

## Iterations and rationale
1. <hypothesis, change, and result>

## Limitations and next steps
<data coverage, sample size, regime sensitivity, or the next safe experiment>
```
