# snip
## A tool for managing code snippets that can be run on the command line

Snip is a librarian for your scripts. It lets you create, edit, and run scripts without having to keep track of where to save them or add them to your path. It also lets you save commands that you have run in your terminal as snips so that you can name them and easily re-run them later.

## Usage

```
snip save <name> <command>
```

Saves `command` as a named snip. The command will be executed using your default shell.

```
snip new [--interp=string|--args=string+] <name>
```

Opens your system editor to create a new snip. If `interp` is given, the snip will be run using that interpreter. If any args are supplied, they will be passed to the interpreter when the snip is run.

```
snip edit <name>
```

Opens your system editor to edit an existing snip.


```
snip run <name> -- [flags and args for snip]
```

Runs a snip. Any flags and args after the `--` will be passed to the interpreter for the snip. Stdin/stdout/stderr will be piped through to the snip's interpreter.

```
snip install
```

Installs helper functions and aliases so that

### Helpers

```
snip-savelast <name>
```

Saves the last command run in your shell as `name`.

### TODO

- Install snip-savelast helper as: `alias snip-savelast='function _snip_save(){ snip save $1 "$(history 2 | head -1 | cut -d" " -f4-)"; };_snip_save'`
