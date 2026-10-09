# Roadmap

## Done (v0.1, October 2026)

- **Journal reading**: boot list (`journalctl --list-boots -o json`, or its table before
  systemd 251), streaming decode of a boot in journalctl's export format, reading only what
  the analysis uses (warnings and errors, the programs the rules name, logins, the last hour):
  a boot filled by a chatty program opens ten times faster; recorded journals of one boot
  (`journalctl -o json`, `.gz` allowed) for tests and `--file`.
- **Verdict**: clean power-off / reboot / shutdown, died asleep (suspended or hibernated),
  kernel panic, unclean, running; time asleep; gap to the next boot.
- **Rules engine**: TOML rules (match, group, noise), named fields, title templates,
  merging (same title or whole rule, within a window), timeline-only context, user rules
  overriding built-ins, self-tests from `[[rule.example]]`.
- **Built-in rules**: memory (OOM report, OOM kill, systemd-oomd), crashes, failed
  services, GPU (amdgpu, i915, NVIDIA Xid), kernel (panic, hung task, soft/hard lockup,
  BUG/Oops), disks (ext4, btrfs, I/O errors, NVMe timeouts), temperature, suspend and
  hibernation, logins and sudo, network, audio, D-Bus, firmware, known harmless issues, noise.
- **Analysis**: findings by severity, "no warnings for N minutes before the end",
  out-of-memory breakdown (processes / shared / GPU memory, swap, biggest processes),
  last hour of lines with repeats folded.
- **Outputs**: `whyd list`, plain-text summary, Markdown report with redaction; summaries of
  ended sessions cached in `~/.cache/whyd`.
- **Interface**: two fixed panels (sessions; the selected session with its time line and
  problems table), arrows only, full-screen views for a problem's journal lines, the last
  minutes and the report, help; terminal colours only; keyboard only; 100×28 minimum.
- **Tests**: on four real recorded boots (an OOM freeze, a quiet unclean end, a clean
  power-off, a reboot with a suspend), every rule's examples, redaction, screen sizes.
- **Release**: CI (gofmt, vet, tests), GoReleaser builds for amd64 and arm64, Arch, Debian
  and RPM packages, `install.sh`.

## Next (v0.2)

- More real journals in `testdata/`: a GPU hang, a died-asleep (battery), a kernel panic via
  pstore, with rules checked against them
- Crash detail: the `coredumpctl info` backtrace and the package that owns the program
- `golangci-lint` in CI
- Redaction of Bluetooth device names

## Ideas for later

None of these changes the core (sources, rules, events, analysis, output): each adds a
source, a rule set or an output.

- **`whyd diff`**: what changed between a good boot and a bad one: packages upgraded in
  between (pacman, apt, dnf, zypper logs), kernel, services, new problems. Many freezes start
  right after a kernel, mesa or driver update.
- **`whyd watch`**: a user service following the live journal and `/proc/pressure/memory`
  that sends a desktop notification when a serious pattern starts (memory pressure for
  minutes, repeated GPU errors, a crash loop, rising temperature).
- **Kernel panics**: panics rarely reach the journal; with pstore enabled, read the panic
  that `systemd-pstore` saves on the next boot and show it as the verdict.
- **Hardware context in the report**: CPU, GPU, driver versions, RAM, the first things
  anyone asks on a forum; opt-in and redacted.
- **Export**: `--json` for scripts.
- **More rules**: per GPU vendor, Wi-Fi and Bluetooth drivers, laptop ACPI quirks,
  compositors, filesystems, SMART and NVMe.

## Out of scope

- Fixing things: whyd explains, it never changes the system.
- Uploading logs anywhere: the report is printed, you decide where it goes.
- Being a general journal viewer: whyd answers one question, why did it stop.

## Rules of thumb

- Real journals beat imagined cases: every bug report should end with a new testdata file.
- A rule without a real example line doesn't go in.
