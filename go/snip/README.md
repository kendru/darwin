# snip
## A tool for managing code snippets that can be run on the command line

Snip is a librarian for your scripts. It lets you create, edit, and run scripts without having to keep track of where to save them or add them to your path. It also lets you save commands that you have run in your terminal as snips so that you can name them and easily re-run them later.

## Usage

```
snip save <name>
```

Saves the last command run as a snip. The command will be executed using your default shell.

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

