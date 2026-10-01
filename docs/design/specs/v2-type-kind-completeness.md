# Spec: v2 `type.kind` completeness — alias & defined (codeanalyzer-go)

> **Single-repo spec.** Lives beside the analyzer it changes
> (`codeanalyzer-go/docs/design/specs/`). A follow-up correction to the v2 L1
> emission designed in [`v2-l1-emission.md`](./v2-l1-emission.md); the `type.kind`
> vocabulary was settled there — this spec fills out the two values the emitter
> never produced.

## Contract-Impact Triage

**Does this change schema v2 output?** Yes — it changes the realized value of the
`type.kind` field for a class of inputs. The *vocabulary* (`struct | interface |
alias | defined`) was already decided in `v2-l1-emission.md` / CLAUDE.md § `type`;
the emitter only ever produced `struct|interface`. This spec corrects the output so
`alias` and `defined` emit for the declaration shapes that warrant them. No new
node/edge kind, no new field, no id-shape change — a leaf-value correction within an
already-designed field.

| Change type | Analyzers | SDKs | Docs |
| --- | --- | --- | --- |
| Schema v2 output correction (`type.kind` completeness) | **codeanalyzer-go** (parser classification + emitter) | none wired (python-sdk has no Go backend yet) | CLAUDE.md note if it claims current partial behavior |

Single-analyzer change. python-sdk has no Go wiring (no `cldk/analysis/go`), so no
SDK model remap is forced; a Go SDK, when it lands, enumerates all four kinds from
the start.

## The defect this corrects

`typeKind` (internal/syntactic_analysis/v2emit/emit.go) returned `interface` if
`IsInterface` else `struct` — full stop. The v1 `GoType` model carried only an
`IsInterface bool`, so the parser never recorded enough to distinguish `alias` or
`defined`. Repro (confirmed 2026-10-01): `type Celsius float64` and
`type MyAlias = X` both emit `kind:"struct"` — a misclassification, not a benign
missing value.

## Design-loop decisions (the transcript)

Run node by node WITH the user, on the `type` node's `kind` field. The four-value
vocabulary was already locked in `v2-l1-emission.md`; the open decisions were the
plumbing and the classification rule.

### Decision 1 — model carrier (LOCKED: `Kind` on `GoType`)

Add `Kind string` to `schema.GoType`, classified in the symbol-table builder (where
the `*ast.TypeSpec` is in hand); retire `IsInterface`; `typeKind` becomes
`return t.Kind`.

- **Why:** classification happens once, in the parser where the AST lives, and the
  v2 emitter stays a *pure re-serializer of v1 facts* (emit.go's stated invariant) —
  it reads a fact, never computes one.
- **Rejected:** composing from `IsAlias`+`IsInterface` booleans (splits the rule
  across parser+emitter, 2 bools encode 4 states so illegal combos become
  representable); re-deriving in the emitter from the AST (couples the emitter to
  `go/ast`, breaking the invariant).
- **Migration note:** `IsInterface` is read by the structural-satisfaction
  (`interfaces[]`) path; those readers switch to `Kind == "interface"`.

### Decision 2 — classification rule (LOCKED: syntactic, by declaration shape)

`kind` is driven by the AST node directly under the `TypeSpec`, with alias taking
priority. Priority order (first match wins):

1. `spec.Assign.IsValid()` (a `type X = Y` alias, regardless of what `Y` is) → **alias**
2. `spec.Type` is `*ast.StructType` literal → **struct**
3. `spec.Type` is `*ast.InterfaceType` literal → **interface**
4. otherwise (`Ident`, `FuncType`, `MapType`, `ArrayType`, `SelectorExpr`, …) → **defined**

- **Why syntactic:** matches the committed vocabulary's wording — `kind` = Go's
  "type-**declaration** shapes", `defined` = `type X Y` over a non-struct (CLAUDE.md's
  own `type Celsius float64` example). `kind` records how a type was *declared*;
  what it *satisfies* already lives in the computed `interfaces[]` field. Keeping the
  two separate avoids `kind` carrying a second, overlapping semantic signal.
- **Alias beats literal:** `type X = struct{...}` is an `alias` — `.Assign` means "X
  is another name for Y", the strongest syntactic signal; it is not a new struct type.
- **`type MyReader io.Reader` → defined** (a distinct named type, not an interface
  declaration), **not** interface. Resolving underlyings via `go/types` was rejected:
  it contradicts the spec, erases the distinct-named-type fact, needs optional
  `TypesInfo` (inconsistent fallback when absent), and duplicates `interfaces[]`.
- **Validated** against the real Go AST (7 cases incl. both corners) before this spec
  was written.

## Implementation plan (the backend rung consumes this)

1. `schema.GoType`: add `Kind string`; remove `IsInterface` (update the `interfaces[]`
   readers to `Kind == "interface"`).
2. `symbol_table.go` `buildType` (~L375): replace the `StructType`/`InterfaceType`
   switch with the 4-way priority classifier (alias-first). Struct-field / embedded /
   interface-method collection still keys off the literal cases.
3. `emit.go` `typeKind`: `return t.Kind`.
4. Gate: extend the L1 gate (gate_test.go) with alias + defined fixtures in
   `testdata/multipackage` (or a dedicated fixture) asserting all four kinds —
   the current gate only exercises `interface` (Processor), which is why the gap
   stayed green.

## Release plan

- **Rides the same v2 train** as the L1/L2 emission work and the determinism fix —
  all pre-first-release of codeanalyzer-go's v2, so no coordinated major bump beyond
  v2.0.0 itself.
- **Behavior-affecting** (output `kind` values change) → forces a release once the
  v2 train ships; not a separate train.
- **Superset-safe:** no v1 fact is dropped; `kind` becomes *more* precise. Types that
  were `struct` and are genuinely structs are unchanged.
- **No SDK pin** to move (no Go wiring in python-sdk yet).

## Decomposition (to confirm with user)

Single repo, lands in one PR → **one issue**, implementation steps as a checklist
under it. (Not a parent/child stack — nothing spans repos or separate clocks.)
