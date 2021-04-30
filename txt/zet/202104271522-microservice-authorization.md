---
tags: ["authorization","microservices","multitenancy"]
created: Tue Apr 27 15:22:00 MDT 2021
---

# Microservice Authorization

One fundamental trade-off is how to structure trust domains. There are three basic approaches that correspond to _network perimeter_, _defense in depth_ and _[public everything](#202104271524)_.

1. *Network perimeter:* Have an API gateway that authenticates requests and applies an authorization policy before forwarding requests to services that implicitly trust them. The API gateway may translate authorization claims to HTTP headers (or similar) if necessary.
2. *Defense in depth:* Authorization happens in multiple layers. First, there is a network layer that is typically secured using a VPN or cloud VPC technology. As in the _gateway security_ model, only a reverse proxy server is exposed in both the private network and the Internet. However, service to service traffic is also authorized using mTLS, OAuth 2/JWT, or some other mechanism. Common authorization concerns can be pushed into a service mesh that also handles mTLS.
3. *Public everything:* Each service accepts traffic on behalf of an end user. Requests happen over HTTP using a text-based protocol like OpenAPI. Common authorization logic is enforced using common libraries and documented organizational standards. Each service's trust domain is the same, and it is the public domain.

_Zero trust_ is a related concept that map apply to either _defense in depth_ or _public everything_. The core assumption is that any request could be from a malicious actor and needs to be authenticated and authorized at every point. The distinction in how this concept is applied to defense in depth and public everything is that the former typically relies on at least transport-level security (e.g. mTLS), whereas the latter relies only on bearer tokens or some other tamper-proof mechanism for proving identity.

## Multitenant Concerns

In a multitenant environment, most requests (excepting certain administrative operations) should have some tenant as part of the request context. If the _gateway security_ model is adopted, then the API Gateway is responsible for authenticating the request and adding a trusted header to communicate the tenant to downstream services. If either of the other models is employed, the tenant is usually included in a session context in a common database or in some cryptographically verifiable token that is passed through with each request. In either case, there has to be a trusted party that either creates the session or signs the token that is passed along.
