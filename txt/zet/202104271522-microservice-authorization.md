---
tags: ["authorization","microservices","multitenancy"]
created: Tue Apr 27 15:22:00 MDT 2021
---

# Microservice Authorization

One fundamental trade-off is how to structure trust domains. There are three basic approaches that correspond to _gateway security_, _defense in depth_ and _[public everything](#202104271524)_.

1. *Gateway security:* Have an API gateway that authenticates requests and applies an authorization policy before forwarding requests to services that implicitly trust them. The API gateway will translate authorization claims to HTTP headers (or similar) if necessary.
2. *Defense in depth:* Authorization happens in multiple layers. First, there is a network layer that is typically secured using a VPN or cloud VPC technology. As in the _gateway security_ model, only a reverse proxy server is exposed in both the private network and the Internet. However, service to service traffic is also authorized using mTLS, OAuth 2/JWT, or some other mechanism. Common authorization concerns can be pushed into a service mesh that also handles mTLS.
3. *Public everything:* Each service accepts traffic on behalf of an end user. Requests happen over HTTP using a text-based protocol like OpenAPI. Common authorization logic is enforced using common libraries and documented organizational standards. Each service's trust domain is the same, and it is the public domain.


## Multitenant Concerns

In a multitenant environment, most requests (excepting certain administrative operations) should have some tenant as part of the request context. If the _gateway security_ model is adopted, then the API Gateway is responsible for authenticating the request and adding a trusted header to communicate the tenant to downstream services. If either of the other models is employed, the tenant is usually included in a session context in a common database or in some cryptographically verifiable token that is passed through with each request. In either case, there has to be a trusted party that either creates the session or signs the token that is passed along.
