# Burling AIP Conformance Action

Run [burling](https://github.com/goweft/burling) — the conformance validator
and IBCT chain auditor for the Agent Identity Protocol (`draft-prakash-aip-00`)
— in CI, and surface findings as [SARIF](https://sarifweb.azurewebsites.net/)
in GitHub code scanning.

The action is a thin Docker wrapper over `burling lint --format sarif`. It
produces a SARIF document and exposes its path as an output; pair it with
[`github/codeql-action/upload-sarif`](https://github.com/github/codeql-action)
to get inline annotations on pull requests and alerts in the Security tab.

## Usage

```yaml
name: AIP conformance
on: [pull_request]

permissions:
  contents: read
  security-events: write   # required by upload-sarif

jobs:
  burling:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - id: burling
        uses: goweft/burling-action@v0
        with:
          token: path/to/token.jwt

      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: ${{ steps.burling.outputs.sarif-file }}
```

By default the action does **not** fail the step on findings — it writes the
SARIF and lets code scanning surface the alerts, so the `upload-sarif` step
always runs. To gate the build directly instead, set `fail-on-error: true`.

## Inputs

| Input           | Required | Default         | Description                                                                                 |
|-----------------|----------|-----------------|---------------------------------------------------------------------------------------------|
| `token`         | yes      | —               | Path to the artifact to validate, relative to the repository root.                          |
| `command`       | no       | `lint`          | burling subcommand: `lint`, `validate`, `validate-identity`, or `audit-chain`.              |
| `strict`        | no       | `false`         | Promote WARNING findings to failures (passes `--strict`).                                    |
| `fail-on-error` | no       | `false`         | Fail this step on an ERROR finding. Leave false so a later `upload-sarif` step still runs.   |
| `output`        | no       | `burling.sarif` | Path to write the SARIF document, relative to the repository root.                           |

## Outputs

| Output       | Description                                                          |
|--------------|---------------------------------------------------------------------|
| `sarif-file` | Path to the generated SARIF file, for `upload-sarif`'s `sarif_file`. |

## What burling checks

burling validates identity documents and compact (JWT/EdDSA) IBCTs against the
mechanically-checkable normative requirements of the AIP draft. Chained mode and
scope attenuation are stubbed in the current release and emit informational
findings. See the [burling conformance matrix](https://github.com/goweft/burling)
for the full list of checks and their severities.

## Versioning

Releases are tagged with semver (`v0.1.0`) plus a moving major alias (`v0`). Pin
the major alias for automatic compatible updates, or pin an exact tag for full
reproducibility. The image builds a pinned burling release, so a given action
tag always runs a known validator version.

## License

Apache License 2.0. See [LICENSE](LICENSE).
