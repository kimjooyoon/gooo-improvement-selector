# v0.1.0 release evidence and immutability audit

`v0.1.0` is preserved as `REFUTED_RELEASE_IMMUTABILITY`. It must not be
deleted, recreated, or counted as an immutable-release success.

The official API audit observed `Release.immutable=false` for the release and
`enabled=false` for the repository immutable-releases setting at audit start.
The setting was then enabled through the official repository endpoint. The
machine-readable record is
`.gooo/release-history-provenance-receipt.json`.

The v0.1.0 tag and assets remain historical evidence:

- tag object: `174b9f975b18953359ca7a0a253662a79f907c4e`
- tag target: `a2ad62caf907e87703abe0c6e8debb0d3602c8d3`
- evidence asset: `gooo-improvement-selector-v0.1.0-evidence.tar.gz`, 89681 bytes, `sha256:972f35d2893aaff6b07369768bbd79f1edaebd7e3c525c194ee6745580659c9d`
- release lock: `release-lock-v0.1.0.json`, 644 bytes, `sha256:d507c0bd4652a3b6496ee5cdd89c44795aeeb292c4b5f5065772b3cac0dec671`
- checksums: `SHA256SUMS`, 206 bytes, `sha256:4064705a2609e1f7f296ea957a27b6f17c79ab75ffab11591dd81be73cdc358c`

The next release is eligible only after repository immutable releases are
enabled, the merged main target has successful CI, and both REST and GraphQL
report `immutable=true` for the new release. The prior v0.1.0 release remains
untouched.
