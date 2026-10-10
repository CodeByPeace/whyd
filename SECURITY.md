# Security

whyd only reads the journal: it never changes the system and never sends anything anywhere.
The part that matters for security is the report, which people paste in public forums.

## What to report privately

- a report (`whyd report`, or `R` in the interface) that shows personal data despite the
  redaction: user names, host names, network names, addresses, serial numbers;
- anything else that could expose data of the person running whyd.

Use [private vulnerability reporting](https://github.com/vmaltarello/whyd/security/advisories/new),
not a public issue: an issue would publish the data again. Include the whyd version and the
redacted line as it came out, with the personal part replaced by something similar but not
real.

You will get an answer within a week. Fixes go into a new release, and the report is
credited unless you prefer otherwise.

## Supported versions

Only the latest release gets fixes.
