# Dora Vota Wasmd dependency fork

Branch: `doravota/sdk055-restricted`.
Base: upstream `v0.70.4`, commit `70c6321f48bfff91dd581a40eeb706a2a24eaf7c`.

This branch extracts the Wasm dependency patches previously held in Dora Vota
`third_party/wasmd-v055-compat` (Dora Vota source commit `f06b375`). It is intended
as the `x/wasm` library used by Dora Vota's SDK 0.55 application. It is not a
standalone SDK 0.55 port of the upstream wasmd CLI/application.

## Local changes

* Add the explicit keeper option `WithDeploymentDisabled()`. The default remains
  upstream behavior; Dora Vota always enables the option in its application.
* When enabled, reject create, instantiate (both forms), and migrate at the shared
  execution boundary, before VM execution or state mutation. Ordinary, privileged
  and nested callers converge at these entry points.
* Existing contract execution/query, genesis import and snapshot restoration
  remain available. The guard does not repair arbitrary existing contract bugs.
  Reopening permissions alone cannot override it; changing the application's
  fixed option requires a coordinated binary upgrade.
* Exclude the obsolete `keeper/test_common.go` helper importing removed SDK
  x/params runtime; remove its two public aliases. This is unchanged from the
  validated local snapshot. Upstream test source files remain in Git, but tests
  relying on that helper need separate SDK 0.55 adaptation and are not claimed
  to pass. This branch does not claim upstream `go test ./...` acceptance.

* Adapt the legacy v2 migration test to use its read-only Subspace interface
  with a separate KV parameter fixture. This removes an obsolete x/params keeper
  dependency while preserving all four parameter migration cases.

The module path stays `github.com/CosmWasm/wasmd`. The consuming application uses
`replace github.com/CosmWasm/wasmd => github.com/DoraFactory/wasmd <fixed version>`.
The application selects SDK 0.55, CometBFT 0.40 and the pinned IBC compatibility
commit, and supplies the historical x/params type compatibility package.
The upstream go.mod is retained to keep this extraction scoped; it alone is not
the tested SDK 0.55 dependency graph. See the consuming application's go.mod.
WasmVM remains official v3.0.8; no VM code is forked here.

## Review and verification

Compare `v0.70.4...doravota/sdk055-restricted`. Pin a full commit in Dora Vota and
record the resulting Go pseudo-version and module checksums. Do not reference a
floating branch in a release.

Validation is performed in Dora Vota: full application Go tests, real WasmVM
contract query/withdrawal and forbidden deployment tests, module verification,
Linux build, fresh-chain Ed25519 block production, restart and export.
Source provenance and actual outcomes are recorded in Dora Vota's
`docs/v1.0.0-preserve/FORK-MIGRATION.md`.

Maintainers must explicitly review/rebase upstream fixes. No upstream release
tag or default branch should be overwritten for this compatibility work.
