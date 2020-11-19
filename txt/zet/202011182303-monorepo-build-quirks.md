---
tags: ["javascript", "yarn"]
created: Wed Nov 18 23:03:41 MST 2020
---

# Monorepo Build Quirks

Monorepos work well for development, but they can make production deployment difficult. For example, in a monorepo, you want to hoist dependencies so that a single instance of some library can be shared by multiple packages. This can even help in the case where a module maintains some state, and if multiple instances of that module are imported, things break. However, for deploying multiple services, the dependencies must be "unhoisted". This can usually be done with a bundler, but there are still odd compatibility issues that need to be kept in mind - such as not being able to statically analyse dynamic imports, which must be included in the bundle explicitly.

Another option is to build and publish all dependent packages to some repository before building and deploying each service. Ideally, some tool would be able to see which packages and services changes, analyse the dependency graph to see what packages or services are downstream, publish all packages in dependency order, and deploy all affected services. Services should have no code-level dependencies on others, so they should always be leaves of the graph that are safe to deploy in parallel.

I ran into an issue today where deployment from a monorepo necessitated duplication. In this case, there was configuration that lived in the root of a repository (specifically `resolutions` in a `package.json` file and `packageExtensions` in a `.yarnrc` file) had to be duplicated in the service's own directory. The version in the root was used during development, and the version in the service directory was used in the Docker container that was built for the service.
