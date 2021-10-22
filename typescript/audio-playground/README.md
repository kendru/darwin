# Audio Playground

This project is a place for developing ideas and strategies for working with DSP for audio synthesis, processing, and visualization.

### Running:

1. Start the server:

```
npm run-script serve
```

2. Start the client:

```
npm run-script watch
```

3. Open your browser to `http://localhost:3000/`

<!--
TODO: Use exo when manifest support lands.
Alternatively, use [exo](https://exo.deref.io/):

```
exo run
```
-->

Note that the file watcher for the client rebuilds code when the source changes, but it does not reload the code in the browser, so a full refresh is needed to pick up changes. The server has no reloading in place, but it likely does not need to change.
