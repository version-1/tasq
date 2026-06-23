# Frontend Agent Notes

Follow the frontend ownership and placement rules in
[docs/design.md](docs/design.md) before adding or moving React components.

Keep `src/app` limited to route entry files, place domain-aware UI under
`src/features`, keep domain-independent primitives under `src/components/ui`,
and keep application shell components under `src/components/layout`.
