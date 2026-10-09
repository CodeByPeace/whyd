# Rules

A rule turns journal lines into an event that a human understands. Rules are TOML files;
the built-in ones live in `rules/` (one file per category), users can add their own in
`~/.config/whyd/rules/`. A user rule with the same `id` replaces the built-in one.

## Your own rules

Rules are text files. whyd reads the built-in ones (in `rules/`, compiled into the program)
and then every `*.toml` file in `~/.config/whyd/rules/`, at every start.

**Add a rule**: create a file there, for example `~/.config/whyd/rules/mine.toml`, with one or
more `[[rule]]` blocks in the format below, each with a new `id`.

**Change a built-in rule**: write a rule with the **same `id`** in your file; yours replaces it
(`whyd rules` lists every id and where it comes from). Typical uses: make something you know is
harmless `severity = "info"`, or hide it with `kind = "noise"`.

**Check your rules**: `whyd rules test` runs the examples of every rule, checks that no earlier
rule takes their lines first, and reports rules that could not be loaded;
`whyd rules test some/file.toml` checks one file before you install it.

What it means in practice:

- **Your rules stay on your machine.** To share one, propose it for `rules/` in the repository,
  with at least one real example line.
- **First match wins.** On each journal line the first rule that recognises it is used.
  Built-in rules come first, so a new rule of yours only sees lines no built-in rule took;
  a rule with a built-in `id` takes that rule's place instead.
- **Give the source.** whyd doesn't read every line of a boot: it reads warnings and errors,
  every line of the programs the rules name in `source`, and everything in the last hour. A
  rule without `source` only sees info and debug lines in that last hour.
- **A wide pattern does damage.** A `match` rule that is too broad turns ordinary lines into
  problems; a `noise` rule that is too broad hides lines that could have mattered.
- **A broken rule is skipped, never fatal.** A syntax error, an invalid pattern or a missing
  field skips that rule (or that file) with a message on start, in the interface and in
  `whyd rules test`; everything else keeps working.

## Format

```toml
[[rule]]
id       = "memory.oom-kill"          # category.name, unique
kind     = "match"                    # match | group | noise
category = "memory"
severity = "critical"                 # critical | warning | info | noise
title    = "out of memory, {process} killed"   # {name} = a named regex group, {source} = who logged it

# Which lines (all conditions must hold)
source   = "kernel"                   # optional: SYSLOG_IDENTIFIER / _COMM
priority = "<=4"                      # optional: syslog priority
message  = 'Out of memory: Killed process (?P<pid>\d+) \((?P<process>[^)]+)\)'

# What to tell the user
explain = """
The kernel ran out of RAM and swap and killed a process to survive.
If memory used by processes looks small, check shared memory and GPU memory
in the OOM report: on integrated GPUs, graphics memory is system RAM.
"""

# Self-test: run by `whyd rules test` (and `go test` for the built-in rules)
[[rule.example]]
match = "Out of memory: Killed process 56703 (Web Content) total-vm:2752116kB"
[[rule.example]]
nomatch = "Out of memory handling enabled"
```

### Merging and context

| Field | Meaning |
|---|---|
| `merge` | `""` (default): occurrences with the same title become one event with a count; `"rule"`: every occurrence of the rule does |
| `merge_window` | only merge occurrences closer than this (`"10m"`); empty = the whole boot |
| `merge_title` | title once merged, may use `{count}`: `"out of memory, {count} processes killed"` |
| `timeline_only` | `true` for context (suspend/resume, a service stopping because its program crashed): kept in the analysis, not listed as a problem |

### Group rules

For problems spread over many lines. Everything between `start` and `end` becomes one event;
fields can be extracted from any line. With a `source`, lines of other programs in between
are left to the other rules. A group also ends after `max_lines` lines (default 300) or
`max_seconds` seconds (default 60): set `max_lines` above the longest real case, an OOM report
lists one line per process.

```toml
[[rule]]
id       = "memory.oom-report"
kind     = "group"
category = "memory"
severity = "critical"
title    = "Out of memory (triggered by {trigger})"
source   = "kernel"
start    = '(?P<trigger>\S+) invoked oom-killer'
end      = 'Out of memory: Killed process'
max_lines = 5000
extract  = [
  'shmem:(?P<shmem>\d+kB)',
  'gpu_active:(?P<gpu>\d+kB)',
]
```

### Noise rules

Hidden from the problems, counted ("hidden: 437 firewall log lines") and folded in the last minutes.

```toml
[[rule]]
id       = "noise.ufw-block"
kind     = "noise"
category = "noise"
severity = "noise"
title    = "Firewall drops"
source   = "kernel"
message  = '^\[UFW BLOCK\]'
```

## Built-in rules

| File | ID | Kind | Severity | Title |
|---|---|---|---|---|
| auth.toml | `auth.failed` | match | warning | {service}: authentication failed |
| auth.toml | `auth.wrong-password` | match | warning | sudo: {n} incorrect password attempts |
| auth.toml | `auth.no-password` | match | info | sudo refused: a password is required |
| auth.toml | `auth.conversation` | match | info | {source}: password prompt ended |
| context.toml | `power.suspend` | match | info | suspended ({mode}) |
| context.toml | `power.resume` | match | info | resumed |
| context.toml | `power.hibernate` | match | info | hibernated |
| context.toml | `power.thaw` | match | info | resumed from hibernation |
| context.toml | `known.rtw88-tx` | match | info | {driver} TX errors, known issue |
| context.toml | `known.portal-gtk` | match | info | portal-gtk failed at login, harmless |
| context.toml | `known.rtw88-dpk` | match | info | {driver} calibration failed (known issue) |
| crash.toml | `crash.coredump` | match | warning | {process} killed by SIG{signal} |
| crash.toml | `crash.coredump-detail` | match | info | core dump details |
| crash.toml | `crash.coredump-lost` | match | info | a core dump could not be saved |
| desktop.toml | `network.wifi-failed` | match | warning | {device}: Wi-Fi connection failed |
| desktop.toml | `audio.underrun` | match | info | audio underruns ({client}) |
| desktop.toml | `dbus.activation-failed` | match | info | D-Bus service {name} could not be started |
| desktop.toml | `dbus.service-file` | match | info | D-Bus service file problem |
| desktop.toml | `network.nm-error` | match | warning | NetworkManager error |
| desktop.toml | `bluetooth.error` | match | info | bluetoothd error |
| firmware.toml | `firmware.boot` | match | info | firmware or hardware setup message |
| firmware.toml | `firmware.tpm` | match | info | TPM NvPCR not available |
| gpu.toml | `gpu.amdgpu-timeout` | match | critical | amdgpu ring {ring} timeout |
| gpu.toml | `gpu.amdgpu-reset` | match | critical | amdgpu GPU reset |
| gpu.toml | `gpu.i915-hang` | match | critical | i915 GPU hang |
| gpu.toml | `gpu.nvidia-xid` | match | critical | NVIDIA Xid {xid} |
| gpu.toml | `gpu.amdgpu-error` | match | warning | amdgpu error in {func} |
| hardware.toml | `disk.ext4-error` | match | critical | ext4 error on {device} |
| hardware.toml | `disk.btrfs-error` | match | critical | btrfs error on {device} |
| hardware.toml | `disk.io-error` | match | critical | I/O error on {device} |
| hardware.toml | `disk.nvme-timeout` | match | critical | NVMe {device} timeout, controller reset |
| hardware.toml | `thermal.critical` | match | critical | critical temperature, emergency shutdown |
| hardware.toml | `disk.fat-dirty` | match | info | {device} not unmounted cleanly (FAT) |
| kernel.toml | `kernel.panic` | match | critical | kernel panic: {reason} |
| kernel.toml | `kernel.hung-task` | match | critical | {process} blocked for over {secs}s |
| kernel.toml | `kernel.soft-lockup` | match | critical | CPU {cpu} soft lockup ({secs}s) |
| kernel.toml | `kernel.hard-lockup` | match | critical | CPU {cpu} hard lockup |
| kernel.toml | `kernel.bug` | match | warning | kernel bug: {what} |
| kernel.toml | `kernel.watchdog-running` | match | info | {dev} still running when closed |
| memory.toml | `memory.oom-report` | group | critical | out of memory, {process} killed |
| memory.toml | `memory.oom-kill` | match | critical | out of memory, {process} killed |
| memory.toml | `memory.oomd-kill` | match | critical | systemd-oomd killed {cgroup} |
| memory.toml | `memory.alloc-trouble` | group | warning | {process}: page allocation {what} |
| noise.toml | `noise.firewall` | noise | noise | firewall log lines |
| noise.toml | `noise.rtw88-push` | noise | noise | rtw88 push errors |
| noise.toml | `noise.wireless-ext` | noise | noise | wireless extensions warnings |
| noise.toml | `noise.wpa-multicast` | noise | noise | wpa_supplicant multicast notices |
| service.toml | `service.dumped` | match | info | {unit} main process crashed |
| service.toml | `service.core-dump` | match | info | {unit} stopped, its program crashed |
| service.toml | `service.failed` | match | warning | {unit} failed: {result} |
| service.toml | `service.start-failed` | match | info | {what} failed to start |
| service.toml | `access.denied` | match | warning | {source}: permission denied |

### Coverage

Measured on 42 real boots of one laptop: of 4,641 error lines (priority 0-3),
57% are recognised by a rule, 42% are known routine noise and 5 lines (0.1%) match no rule; of
10,820 warning lines, 39% / 50% / 10%. Lines that match no rule are never dropped: whyd counts
them and lists them grouped by message (`W` in the interface, "Not recognised" in the
report), which is also the easiest place to find the next rule to write.

## Writing a good rule

- **Match the stable part** of the message, not numbers or device names.
- **Name the groups** you want in the title or the report.
- **The explanation is the product.** Two or three plain sentences of facts: what the
  message means and what it can lead to. No guessed causes, no links and nothing specific to
  one distribution: whyd runs anywhere with systemd.
- **Add examples**: at least one line that matches and one that looks similar but must not.
- **Severity**: critical = can explain a crash/freeze by itself; warning = worth a look;
  info = context and known harmless issues.
