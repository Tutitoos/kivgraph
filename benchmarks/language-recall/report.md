# Recall on real code, by language

How many of the files that depend on a declaration does `find_references` report? The ground truth is the language's own type checker: rename the declaration, type-check again, and the files that gain an error are the files that use it.

- Command: `go run ./benchmarks/language-recall -kivgraph <binary> -repo-root .`
- Date: 2026-11-18, seed 1, up to 100 declarations per language
- Kivgraph `0.10.0` on darwin/arm64, go1.26.11, Version 5.9.3, node v25.9.0

## Corpus

|language|repository|commit|
|---|---|---|
|go|[cobra](https://github.com/spf13/cobra)|`adbc8813901b`|
|typescript|[ky](https://github.com/sindresorhus/ky)|`3541888878c4`|

## Result

Intervals are 95 % percentile intervals from resampling declarations.

|language|configuration|declarations measured|with dependents|file recall|file precision|fully recalled|
|---|---|---|---|---|---|---|
|go|default|100|96|65.9 % (59.4 to 74.6)|100.0 % (100.0 to 100.0)|57.3 % (47.3 to 66.7)|
|typescript|default|100|85|92.7 % (86.6 to 97.4)|90.3 % (83.4 to 96.2)|90.6 % (82.5 to 95.2)|
|go|go.include_tests|100|96|99.2 % (97.8 to 100.0)|100.0 % (100.0 to 100.0)|97.9 % (92.7 to 99.4)|

`file recall` pools the dependent files of every declaration. `fully recalled` is the share of declarations with dependents for which no file was missed (Wilson interval). `go.include_tests` is the default configuration with `go.include_tests: true`; the two Go rows draw the same declarations.

## go (default)

- Declarations listed 443, eligible for the rename 442, measured 100 (oracle failed on 0, Kivgraph refused 0).
- Recall on non-test files: 100.0 % (100.0 to 100.0). On test files: 0.0 % (0.0 to 0.0).
- Declarations with no dependents for which Kivgraph still reports a file: 0 of 4.
- Completeness verdicts: COMPLETE 96, LOWER_BOUND 4.
- Answers that said `COMPLETE` and missed at least one non-test file: 0 of 96.
- Not sampled: name not found in declaring position: 1; outline refused: 1.

|kind|declarations with dependents|file recall|
|---|---|---|
|alias|2|71.4 % (66.7 to 75.0)|
|const|18|79.5 % (68.6 to 90.9)|
|field|18|55.6 % (44.7 to 77.1)|
|func|17|82.6 % (65.0 to 95.7)|
|method|15|57.4 % (40.0 to 70.2)|
|type|15|60.9 % (52.7 to 84.6)|
|var|11|78.9 % (65.0 to 94.4)|

Declarations with a missed file (first 15 of 41):

|declaration|kind|missed|verdict|
|---|---|---|---|
|`Completion` (completions.go:136)|alias|completions_test.go|COMPLETE|
|`BashCompSubdirsInDir` (bash_completions.go:33)|const|bash_completions_test.go, completions_test.go|COMPLETE|
|`ShellCompDirective` (completions.go:45)|type|active_help_test.go, completions_test.go|COMPLETE|
|`CompletionFunc` (completions.go:139)|alias|completions_test.go|COMPLETE|
|`FParseErrWhitelist` (command.go:42)|type|command_test.go|COMPLETE|
|`EnableTraverseRunHooks` (cobra.go:66)|var|command_test.go|COMPLETE|
|`DisableDescriptions` (completions.go:115)|field|completions_test.go|COMPLETE|
|`Group` (command.go:45)|type|command_test.go|COMPLETE|
|`defaultCaseInsensitive` (cobra.go:48)|const|command_test.go|COMPLETE|
|`Args` (command.go:93)|field|args_test.go, bash_completions_test.go, command_test.go, completions_test.go, fish_completions_test.go|LOWER_BOUND|
|`GenYamlTree` (doc/yaml_docs.go:53)|func|doc/yaml_docs_test.go|COMPLETE|
|`Traverse` (command.go:821)|method|command_test.go|COMPLETE|
|`ShellCompDirectiveNoFileComp` (completions.go:66)|const|active_help_test.go, completions_test.go|COMPLETE|
|`defaultCommandSorting` (cobra.go:47)|const|command_test.go|COMPLETE|
|`SuggestFor` (command.go:71)|field|bash_completions_test.go, command_test.go, doc/cmd_test.go|COMPLETE|

## typescript (default)

- Declarations listed 1062, eligible for the rename 398, measured 100 (oracle failed on 0, Kivgraph refused 0).
- Recall on non-test files: 92.7 % (86.8 to 97.5). On test files: n/a.
- Recall on ES private names (`#x`): 100.0 % (100.0 to 100.0) over 13 declarations. On every other declaration: 92.0 % (85.6 to 96.9).
- Declarations with no dependents for which Kivgraph still reports a file: 4 of 15.
- Completeness verdicts: COMPLETE 100.
- Answers that said `COMPLETE` and missed at least one non-test file: 8 of 100.
- Not sampled: kind not sampled: 391; name not found in declaring position: 273; outline refused: 21.

|kind|declarations with dependents|file recall|
|---|---|---|
|class|10|100.0 % (100.0 to 100.0)|
|function|9|100.0 % (100.0 to 100.0)|
|interface|2|100.0 % (100.0 to 100.0)|
|method|13|100.0 % (100.0 to 100.0)|
|property|13|47.6 % (30.0 to 70.0)|
|type|19|100.0 % (100.0 to 100.0)|
|variable|19|100.0 % (100.0 to 100.0)|

Declarations with a missed file (first 15 of 8):

|declaration|kind|missed|verdict|
|---|---|---|---|
|`beforeRetry` (source/types/hooks.ts:301)|property|source/utils/merge.ts|COMPLETE|
|`maxResponseSize` (source/types/options.ts:268)|property|source/core/constants.ts|COMPLETE|
|`throwHttpErrors` (source/types/options.ts:534)|property|source/core/Ky.ts|COMPLETE|
|`onDownloadProgress` (source/types/options.ts:317)|property|source/core/constants.ts, source/types/options.ts, source/utils/body.ts|COMPLETE|
|`error` (source/types/retry.ts:13)|property|source/core/Ky.ts|COMPLETE|
|`baseUrl` (source/types/options.ts:151)|property|source/core/constants.ts, source/types/options.ts|COMPLETE|
|`request` (source/types/hooks.ts:33)|property|source/core/Ky.ts|COMPLETE|
|`request` (source/types/hooks.ts:45)|property|source/core/Ky.ts|COMPLETE|

Declarations with a reported file the compiler did not flag (first 15 of 10). These are unaudited: a use the compiler does not reject, such as one in a comment, is not a false positive.

|declaration|kind|extra|
|---|---|---|
|`Ky` (source/core/Ky.ts:348)|class|source/core/constants.ts, source/errors/ForceRetryError.ts|
|`value` (source/utils/body.ts:149)|method|source/utils/body.ts|
|`SchemaValidationError` (source/errors/SchemaValidationError.ts:25)|class|source/types/standard-schema.ts|
|`delay` (source/utils/delay.ts:9)|function|source/core/Ky.ts|
|`ResponseSizeError` (source/errors/ResponseSizeError.ts:20)|class|source/core/Ky.ts, source/types/hooks.ts, source/types/ky.ts, source/types/options.ts, source/utils/merge.ts|
|`isRawNetworkError` (source/utils/is-network-error.ts:15)|function|source/core/Ky.ts|
|`HTTPError` (source/errors/HTTPError.ts:15)|class|source/types/hooks.ts|
|`ShouldRetryState` (source/types/retry.ts:9)|type|source/index.ts|
|`timeout` (source/utils/timeout.ts:11)|function|source/core/Ky.ts|
|`KyRequest` (source/types/request.ts:1)|type|source/types/hooks.ts|

## go (go.include_tests)

- Declarations listed 443, eligible for the rename 442, measured 100 (oracle failed on 0, Kivgraph refused 0).
- Recall on non-test files: 100.0 % (100.0 to 100.0). On test files: 97.7 % (93.2 to 100.0).
- Declarations with no dependents for which Kivgraph still reports a file: 0 of 4.
- Completeness verdicts: COMPLETE 96, LOWER_BOUND 4.
- Answers that said `COMPLETE` and missed at least one non-test file: 0 of 96.
- Not sampled: name not found in declaring position: 1; outline refused: 1.

|kind|declarations with dependents|file recall|
|---|---|---|
|alias|2|85.7 % (66.7 to 100.0)|
|const|18|100.0 % (100.0 to 100.0)|
|field|18|100.0 % (100.0 to 100.0)|
|func|17|100.0 % (100.0 to 100.0)|
|method|15|100.0 % (100.0 to 100.0)|
|type|15|98.4 % (91.2 to 100.0)|
|var|11|100.0 % (100.0 to 100.0)|

Declarations with a missed file (first 15 of 2):

|declaration|kind|missed|verdict|
|---|---|---|---|
|`CompletionFunc` (completions.go:139)|alias|completions_test.go|COMPLETE|
|`SliceValue` (completions.go:311)|type|completions_test.go|COMPLETE|

## What this does not show

- The oracle and Kivgraph share each language's type checker. Go is checked with `go/packages` and `go/types` on both sides, TypeScript with the repository's own `tsc` on one side and the same checker inside the worker on the other. Type resolution is not what this measures. What it measures is the rest: which files and build variants are loaded, how resolved uses become edges, and what is filtered on the way out.
- The unit is the file, not the reference. A rename makes the compiler reject some uses and not others (after a failed import, the uses of the imported name are not reported), so reference-level ground truth would differ by language. A file is a dependent if it gains an error.
- Implementation and inheritance relations are left out of the truth: an error that says a type no longer satisfies an interface or a base class is not a use of the renamed name, and `find_references` is not the tool for it.
- The sample frame is what Kivgraph lists for each file, so a declaration Kivgraph never recorded cannot be sampled. A recall measured this way is conditional on the declaration being known.
- Only declarations whose name appears in its declaring position on its first line are renamed, so the rename is exact. Declarations written another way are counted as not sampled.
- Go declarations are sampled from non-test files; references from `_test.go` files are part of the truth and are reported apart, because Kivgraph reads test files only with `go.include_tests: true` and this run used the default.
- One repository per language: the figures describe these repositories at these commits, not the language. A second repository would show how much they move.
- A declaration with no dependents says nothing about recall. It is reported through the share of such declarations for which Kivgraph still reports a reference.
- The TypeScript dependencies are installed with `npm install` and are not pinned by a lockfile, so a rerun can resolve other versions.
