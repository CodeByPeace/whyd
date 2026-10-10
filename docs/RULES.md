# Rules

whyd recognises problems in the journal with **rules**: short blocks of text that say "a line
like this one means this problem". The built-in rules come with whyd; you can write your own
for the messages it doesn't know yet, without touching any code.

This page walks you through writing one, step by step, then goes further and ends with a
reference of every field.

- [How rules work](#how-rules-work)
- [Your first rule, step by step](#your-first-rule-step-by-step)
- [Going further](#going-further)
- [Reference](#reference)
- [Writing a good rule](#writing-a-good-rule)
- [Built-in rules](#built-in-rules)

## How rules work

whyd reads the journal of a session line by line, and tries the rules on each line, in order.
The first rule that recognises the line takes it, and then:

- a **match** rule turns it into a problem in the list, with a level, a title and an
  explanation;
- a **noise** rule hides it and only counts it ("437 firewall log lines");
- if no rule takes it and the line is a warning or an error, whyd lists it as **not
  recognised** (`W` in the interface). Those are the lines a new rule can explain.

## Your first rule, step by step

We'll write a rule for a real message whyd doesn't recognise yet.

### 1. Find the line

Open whyd, choose a session and press `W`: these are the warnings and errors no rule took.

```
┌─ Not recognised: 1 errors, 8 warnings, 9 messages ──────────────────────────────────────┐
│ LEVEL  TIMES  FIRST      SOURCE          MESSAGE                                        │
│ warn   1      13:45      (udev-worker)   ACAD: Process '/bin/sh -c 'echo 75 80 > /sys/… │
│ warn   1      13:45      (udev-worker)   cfg80211: Process '/usr/bin/set-wireless-regd… │
```

Select a line to read it in full:

```
cfg80211: Process '/usr/bin/set-wireless-regdom' failed with exit code 1.
```

Two things matter: the **message** (the text) and the **source**, the program that logged it,
here `(udev-worker)`. The message says that a command started by a udev rule failed.

### 2. Write a rule that recognises it

Rules go in TOML files in `~/.config/whyd/rules/`; any name ending in `.toml` will do. Create
`~/.config/whyd/rules/mine.toml` with:

```toml
[[rule]]
id       = "mine.udev-command"
kind     = "match"
category = "hardware"
severity = "warning"
title    = "a udev command failed"
message  = 'failed with exit code'
```

Line by line:

- `[[rule]]` starts a rule. A file can hold as many as you like, one after the other.
- `id` names the rule. Any text, unique; by habit `something.what-it-is`.
- `kind = "match"`: each line it recognises is one occurrence of a problem.
- `category`: one word for the area: `hardware`, `network`, `disk`…
- `severity`: how serious it is: `"critical"`, `"warning"` or `"info"`, shown as `err`,
  `warn`, `info`.
- `title`: what the problem is called in the list.
- `message`: the text to look for in the line. Note the single quotes `'...'`.

Run `whyd --text` (or open whyd) on that session:

```
TIME          LEVEL  WHAT                                                 TIMES
Wed 13:45     warn   a udev command failed                                2
```

Both lines are recognised, and since they have the same title they are one problem that
happened 2 times.

### 3. Say which program logs it

"failed with exit code" could come from any program. Add `source` to limit the rule to one:

```toml
source   = "(udev-worker)"
```

Write it exactly as in the SOURCE column, parentheses and capitals included. Kernel messages
have the source `kernel`.

Always give the source when you know it: whyd doesn't read every line of a long session, but
it does read every line of the programs named in a `source` (more in [What whyd
reads](#what-whyd-reads)).

### 4. Show what's in the line

The title would be more useful with the device and the command. Parts of the message can be
taken out as **fields** and put in the title. Change `message` and `title` to:

```toml
title    = "{device}: udev command failed"
message  = '^(?P<device>\S+): Process (?P<command>.+) failed with exit code (?P<code>\d+)'
```

`message` is a regular expression: text, plus a few symbols that stand for "any word",
"any number" and so on. Here:

| Piece | Means |
|---|---|
| `^` | the message starts here |
| `\S+` | a word: anything up to the next space |
| `.+` | anything |
| `\d+` | a number |
| `(?P<device>...)` | keep what is inside as the field `device` |

So `(?P<device>\S+)` is "a word, kept as `device`", and `{device}` in the title puts it back:

```
TIME          LEVEL  WHAT                                                 TIMES
Wed 13:45     warn   ACAD: udev command failed                            1
Wed 13:45     warn   cfg80211: udev command failed                        1
```

The titles are now different, so they are two problems. `{source}` is always available too.
[Patterns](#patterns) lists more of the symbols.

### 5. Explain it

Two more fields say what the problem means. Add them to the rule:

```toml
note     = "{command} exited with code {code}"
explain  = """
A udev rule ran a command when the device appeared or changed, and the command failed.
What the rule was there to do, such as a setting or a permission, was not done.
"""
```

`note` is one short line; `explain` is a few sentences, and the `"""` let it span several
lines. This is where each field appears in whyd:

```
┌─ Problems: 4 warn, 6 info ─────────────────────────────────────┐
│ TIME       LEVEL  WHAT                                   TIMES │
│ 13:45      warn   ACAD: udev command failed                  1 │
│ 13:45      warn   cfg80211: udev command failed              1 │  ← severity, title, times
│ 13:45      warn   NetworkManager error                       1 │
│                                                                │
│ '/usr/bin/set-wireless-regdom' exited with code 1              │  ← note
│                                                                │
│ A udev rule ran a command when the device appeared or          │  ← explain
│ changed, and the command failed. What the rule was there to    │
│ do, such as a setting or a permission, was not done.           │
└────────────────────────────────────────────────────────────────┘
```

The same title, level and note go into `whyd --text` and into the report.

### 6. Test it

Add the real line as an example, and a similar line the rule must leave alone:

```toml
[[rule.example]]
match = "cfg80211: Process '/usr/bin/set-wireless-regdom' failed with exit code 1."
[[rule.example]]
nomatch = "cfg80211: Process '/usr/bin/set-wireless-regdom' succeeded."
```

Then run:

```
$ whyd rules test
ok: 53 rules, 68 examples
```

`whyd rules test` checks every rule: mistakes in the file, examples that don't do what they
say, and rules that can never fire. If something is wrong it tells you what:

```
SKIP /home/you/.config/whyd/rules/mine.toml: rule "mine.udev-command" skipped: unknown field "sorce", did you mean "source"?
```

A rule with a mistake is skipped; whyd keeps working with the others.

### The whole rule

```toml
[[rule]]
id       = "mine.udev-command"
kind     = "match"
category = "hardware"
severity = "warning"
title    = "{device}: udev command failed"
note     = "{command} exited with code {code}"
source   = "(udev-worker)"
message  = '^(?P<device>\S+): Process (?P<command>.+) failed with exit code (?P<code>\d+)'
explain  = """
A udev rule ran a command when the device appeared or changed, and the command failed.
What the rule was there to do, such as a setting or a permission, was not done.
"""

[[rule.example]]
match = "cfg80211: Process '/usr/bin/set-wireless-regdom' failed with exit code 1."
[[rule.example]]
nomatch = "cfg80211: Process '/usr/bin/set-wireless-regdom' succeeded."
```

Rules are read every time whyd starts: save the file and run whyd again.

## Going further

### Hide routine lines: noise

Some lines are harmless but so frequent that they bury everything else. A **noise** rule
hides them and counts them ("hidden: 437 firewall log lines"); in the last minutes they are
folded, and `N` shows them.

```toml
[[rule]]
id      = "noise.firewall"
kind    = "noise"
title   = "firewall log lines"
source  = "kernel"
message = '\bIN=\S* OUT=\S* .*\bSRC='

[[rule.example]]
match = "[UFW BLOCK] IN=wlan0 OUT= MAC=02:1a:2b:3c:4d:5e SRC=192.168.50.254 DST=192.168.50.94 LEN=468"
```

A noise rule needs no `severity` or `category`: its lines are only counted. Use it sparingly,
a hidden line is a line nobody reads. For something harmless but worth knowing, prefer a
`match` rule with `severity = "info"`.

### Problems on several lines: group

Some problems are reported over many lines: an out-of-memory report is a header, the state
of the memory, one line per process and the process killed. A **group** rule makes one
occurrence of the whole block:

```toml
[[rule]]
id        = "memory.oom-report"
kind      = "group"
category  = "memory"
severity  = "critical"
title     = "out of memory, {process} killed"
source    = "kernel"
start     = '(?P<trigger>\S+) invoked oom-killer'
end       = 'Out of memory: Killed process (?P<pid>\d+) \((?P<process>[^)]+)\)'
max_lines = 5000

[[rule.example]]
match = "NetworkManager invoked oom-killer: gfp_mask=0x140cca(GFP_HIGHUSER_MOVABLE|__GFP_COMP), order=0, oom_score_adj=0"
```

- The block **starts** at a line matching `start`, takes the next lines of the same source,
  and **ends** at a line matching `end`, which is part of it.
- It also ends after `max_lines` lines (300 if not given), when a line comes more than
  `max_seconds` after the start (60 if not given), or when another `start` line begins a new
  block. Then `end` was never seen and its fields are missing: give `max_lines` more than the
  longest real block.
- Fields come from `start` and `end`, and from any line of the block with
  `extract = ['shmem:(?P<shmem>\d+kB)']` (a list of patterns; the first value found wins).
- `Enter` on the problem in whyd shows every line of the block.
- In a group rule, `start` and `end` replace `message`, and the examples are tested on
  `start`.

### Many occurrences, one problem: merge

By default, occurrences with the same title are one problem, counted in TIMES, over the whole
session. Three fields change that:

```toml
merge        = "rule"
merge_window = "10m"
merge_title  = "out of memory, {count} processes killed"
```

- `merge = "rule"`: every occurrence of the rule joins the same problem, even with different
  titles. Leave it out to merge by title.
- `merge_window`: only join an occurrence that comes within this time of the previous one
  (`"30s"`, `"10m"`, `"1h"`); later ones start a new problem.
- `merge_title`: the title once two or more occurrences are joined; `{count}` is how many.

### Context, not a problem: timeline_only

`timeline_only = true` makes a rule take its lines without listing a problem. They are not
"not recognised" anymore, and they are marked `*` in the last minutes. It is for lines that
only give context (the system suspended) or that repeat another problem (systemd reporting
the crash that systemd-coredump already reported).

### Only the serious lines: priority

Every journal line has a priority, from 0 (the most serious) to 7:

| 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
|---|---|---|---|---|---|---|---|
| emerg | alert | crit | err | warning | notice | info | debug |

`priority = "<=4"` limits a rule to warnings and worse, `"<=3"` to errors and worse. Useful
when a program logs the same words in harmless lines too.

### Change or silence a built-in rule

Write a rule with the **same `id`** as a built-in one (`whyd rules` lists them all): yours
replaces it. It replaces it whole, nothing is kept from the built-in one, so copy the rule
from [`rules/`](../rules) and change what you need. `whyd rules` then shows `(replaces
built-in)` next to it.

- **Quieter**: the same rule with `severity = "info"`.
- **Hidden**: the same `id`, `kind = "noise"`, its `title` and its `message`.

## Reference

### Fields

Every rule:

| Field | Required | Values |
|---|---|---|
| `id` | yes | any text, unique; a rule with the `id` of an earlier one replaces it |
| `kind` | yes | `"match"`, `"group"`, `"noise"` |
| `title` | yes | text, with `{fields}` |
| `category` | yes, except noise | one word: `"memory"`, `"gpu"`, `"network"`… whyd's "before" line names at most one problem per category |
| `severity` | yes, except noise | `"critical"` (shown as `err`): can explain a crash or a freeze by itself; `"warning"` (`warn`): worth a look; `"info"`: context, known harmless messages. Noise rules: `"noise"` or nothing |
| `source` | no | the program, exactly as whyd shows it; kernel messages: `"kernel"` |
| `priority` | no | `"<=N"`, N from 0 to 7 |
| `[[rule.example]]` | recommended | `match = "..."` or `nomatch = "..."`, one per block, as many blocks as you like |

Recognising lines:

| Field | Kind | Values |
|---|---|---|
| `message` | match, noise (required) | a pattern |
| `start`, `end` | group (required) | patterns |
| `extract` | group | a list of patterns: `['a(?P<x>\d+)', 'b(?P<y>\w+)']` |
| `max_lines` | group | a number, 300 if not given |
| `max_seconds` | group | a number, 60 if not given |

What whyd shows (not for noise rules):

| Field | Values |
|---|---|
| `note` | one short line, with `{fields}`; repeated problems show "from 10:08 to 11:35" instead |
| `explain` | sentences; a blank line starts a new paragraph; whyd shows the whole sentences that fit |
| `merge` | `"rule"`, or nothing to merge by title |
| `merge_window` | a duration: `"30s"`, `"10m"`, `"1h"`; nothing: the whole session |
| `merge_title` | text with `{count}` and `{fields}` |
| `timeline_only` | `true` or `false` (the default) |

A field whyd doesn't know, a field that has no effect for the kind, a wrong value or a
`{field}` no pattern defines is a mistake: whyd skips the rule and says what to fix.

### Patterns

`message`, `start`, `end` and `extract` are regular expressions in [Go's
syntax](https://pkg.go.dev/regexp/syntax), tested on the message of a line (without its
time and source).

| Write | To find |
|---|---|
| `Out of memory` | this text, anywhere in the message; capitals matter |
| `^Failed` / `done\.$` | at the start / at the end of the message |
| `(?i)permission denied` | without caring about capitals |
| `\d+` · `\S+` · `\w+` · `.+` | a number · a word up to a space · letters and digits · anything |
| `[^)]+` | anything up to a `)` |
| `(?:BLOCK\|ALLOW)` | one text or the other |
| `\(` `\[` `\.` `\*` | the characters `(` `[` `.` `*` themselves |
| `(?P<name>...)` | the same, kept as the field `name` |

- Write patterns between single quotes `'...'`: a backslash is then just a backslash. Between
  double quotes it must be doubled (`"\\d+"`).
- Leave out what changes from one time to the next (numbers, devices, addresses), or capture
  it as a field.
- Not available in Go: `(?=...)`, `(?!...)` and `\1`.
- A message can hold several lines (stack traces): `^` and `$` are its start and end, and `.`
  doesn't cross a line break unless the pattern starts with `(?s)`.

### Titles

`title`, `note` and `merge_title` can use `{name}` for a field of the patterns, `{source}` for
the program, and `{count}` (only in `merge_title`) for how many occurrences were merged. Keep
title fields required in the pattern: an optional group that didn't match stays as `{name}`.

### Order

The built-in files in alphabetical order (`auth.toml`, `context.toml`, …), each from top to
bottom, then your files in alphabetical order. On each line the first rule whose `source`,
`priority` and pattern all fit takes it, so your new rules see only the lines no built-in rule
took. A block that a group rule is collecting gets its lines before any other rule.

### What whyd reads

To open a session quickly, whyd doesn't read every line: it reads the warnings and errors,
every line of the programs named in some `source`, the logins, and everything in the last hour
before the end. A rule without `source` therefore sees info and debug lines only in that last
hour.

### Files

- Your rules: `~/.config/whyd/rules/*.toml` (`$XDG_CONFIG_HOME/whyd/rules/` if set). With
  `sudo`, whyd usually reads root's directory instead.
- The built-in rules: [`rules/`](../rules), compiled into whyd.
- `whyd rules` lists every rule, its level, its kind and the file it comes from.
- `whyd rules test` checks your rules; `whyd rules test file.toml` checks one file before
  you put it in place. It exits with status 1 when something is wrong.

## Writing a good rule

- **Start from a real line**, and keep it as a `match` example. Add a `nomatch` example for
  the closest line that must not be taken.
- **Give the source**, and a `priority` when the program writes the same words in harmless
  lines.
- **Match the part that doesn't change**, and capture what you want to show.
- **Explain with facts**: two or three plain sentences on what the message means and what it
  can lead to. No guessed causes, no links, nothing specific to one distribution.
- **Choose the level honestly**: `critical` only for what can stop or freeze the system by
  itself.
- **Share it**: if it helped you, propose it for [`rules/`](../rules) with its real line.

## Built-in rules

Level is what whyd shows: `err`, `warn`, `info`; `context` rules take their lines without
listing a problem; `hidden` lines are counted as noise.

| ID | File | Kind | Level | Title |
|---|---|---|---|---|
| `auth.failed` | auth.toml | match | warn | {service}: authentication failed |
| `auth.wrong-password` | auth.toml | match | warn | sudo: {n} incorrect password attempts |
| `auth.no-password` | auth.toml | match | info | sudo refused: a password is required |
| `auth.conversation` | auth.toml | match | info, context | {source}: password prompt ended |
| `power.suspend` | context.toml | match | info, context | suspended ({mode}) |
| `power.resume` | context.toml | match | info, context | resumed |
| `power.hibernate` | context.toml | match | info, context | hibernated |
| `power.thaw` | context.toml | match | info, context | resumed from hibernation |
| `known.rtw88-tx` | context.toml | match | info | {driver} TX errors, known issue |
| `known.portal-gtk` | context.toml | match | info | portal-gtk failed at login, harmless |
| `known.rtw88-dpk` | context.toml | match | info | {driver} calibration failed (known issue) |
| `crash.coredump` | crash.toml | match | warn | {process} killed by SIG{signal} |
| `crash.coredump-detail` | crash.toml | match | info, context | core dump details |
| `crash.coredump-lost` | crash.toml | match | info | a core dump could not be saved |
| `network.wifi-failed` | desktop.toml | match | warn | {device}: Wi-Fi connection failed |
| `audio.underrun` | desktop.toml | match | info | audio underruns ({client}) |
| `dbus.activation-failed` | desktop.toml | match | info | D-Bus service {name} could not be started |
| `dbus.service-file` | desktop.toml | match | info | D-Bus service file problem |
| `network.nm-error` | desktop.toml | match | warn | NetworkManager error |
| `bluetooth.error` | desktop.toml | match | info | bluetoothd error |
| `firmware.boot` | firmware.toml | match | info | firmware or hardware setup message |
| `firmware.tpm` | firmware.toml | match | info | TPM NvPCR not available |
| `gpu.amdgpu-timeout` | gpu.toml | match | err | amdgpu ring {ring} timeout |
| `gpu.amdgpu-reset` | gpu.toml | match | err | amdgpu GPU reset |
| `gpu.i915-hang` | gpu.toml | match | err | i915 GPU hang |
| `gpu.nvidia-xid` | gpu.toml | match | err | NVIDIA Xid {xid} |
| `gpu.amdgpu-error` | gpu.toml | match | warn | amdgpu error in {func} |
| `disk.ext4-error` | hardware.toml | match | err | ext4 error on {device} |
| `disk.btrfs-error` | hardware.toml | match | err | btrfs error on {device} |
| `disk.io-error` | hardware.toml | match | err | I/O error on {device} |
| `disk.nvme-timeout` | hardware.toml | match | err | NVMe {device} timeout, controller reset |
| `thermal.critical` | hardware.toml | match | err | critical temperature, emergency shutdown |
| `disk.fat-dirty` | hardware.toml | match | info | {device} not unmounted cleanly (FAT) |
| `kernel.panic` | kernel.toml | match | err | kernel panic: {reason} |
| `kernel.hung-task` | kernel.toml | match | err | {process} blocked for over {secs}s |
| `kernel.soft-lockup` | kernel.toml | match | err | CPU {cpu} soft lockup ({secs}s) |
| `kernel.hard-lockup` | kernel.toml | match | err | CPU {cpu} hard lockup |
| `kernel.bug` | kernel.toml | match | warn | kernel bug: {what} |
| `kernel.watchdog-running` | kernel.toml | match | info | {dev} still running when closed |
| `memory.oom-report` | memory.toml | group | err | out of memory, {process} killed |
| `memory.oom-kill` | memory.toml | match | err | out of memory, {process} killed |
| `memory.oomd-kill` | memory.toml | match | err | systemd-oomd killed {cgroup} |
| `memory.alloc-trouble` | memory.toml | group | warn | {process}: page allocation {what} |
| `noise.firewall` | noise.toml | noise | hidden | firewall log lines |
| `noise.rtw88-push` | noise.toml | noise | hidden | rtw88 push errors |
| `noise.wireless-ext` | noise.toml | noise | hidden | wireless extensions warnings |
| `noise.wpa-multicast` | noise.toml | noise | hidden | wpa_supplicant multicast notices |
| `service.dumped` | service.toml | match | info, context | {unit} main process crashed |
| `service.core-dump` | service.toml | match | info, context | {unit} stopped, its program crashed |
| `service.failed` | service.toml | match | warn | {unit} failed: {result} |
| `service.start-failed` | service.toml | match | info, context | {what} failed to start |
| `access.denied` | service.toml | match | warn | {source}: permission denied |

### How much they recognise

On the 42 sessions of the laptop whyd was developed on, this is where each warning and error
line ended up:

| Lines | Recognised by a rule | Hidden as noise | Not recognised |
|---|---|---|---|
| Errors (priority 0 to 3): 4,645 | 2,675 (58%) | 1,964 (42%) | 6 (0.1%) |
| Warnings (priority 4): 10,828 | 4,238 (39%) | 5,440 (50%) | 1,150 (11%) |

Another machine, with other hardware and other programs, will have more lines not recognised:
each of them is a rule waiting to be written.
