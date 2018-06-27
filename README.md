traefik-cert
============

A service for serving certs obtained by traefik.

I originally created this in order to use traefik to managed certs from Let's
Encrypt, and use those certs on other services, like SMTP and IMAP.


Usage
-----

1. Generate a JWT keypair to use for authorizing clients.
2. Deploy the service somewhere so it can access the same certs as traefik.
3. Generate JWT tokens for other services to obtain certs.
4. Use the `getcert` verb to get a cert and save it in a place for the service.


Deployment of service
---------------------

Use the public key from the JWT keypair.

With `docker run`:
```
docker run -v /tmp/acme:/acme \
--label=traefik.frontend.rule=Host:dev.sprinkle.cloud \
--name cert \
-v ${PWD}/public.key:/public.key \
brimstone/traefik-cert
```

As a Docker swarm service:
```
TODO
```


Usage of client
---------------

1. Generate a token with this keypair with a top level `cert` object. This cert
   object needs to have an array `domains` with all of the specific domains.
2. Use the `getcert` verb with the token to connect to the service and request
   a cert.

### Example JWT

Header:
```
{
  "alg": "RS256",
  "kid": "",
  "typ": "JWT"
}
```
Payload:
```
{
  "exp": 1561597620,
  "iat": 1530061620,
  "nbf": 1530061620,
  "cert": {
    "domains": [
      "dev.sprinkle.cloud"
    ]
  }
}
```

### Example usage of `getcert`
```
traefik-cert getcert -u cert.sprinkle.cloud -d mail.sprinkle.cloud -j eyJhbGciOiJSUzI1NiIsImtpZCI6IiIsInR5cCI6IkpXVCJ9…
```


Requirements/Prerequisites
--------------------------

* Go
* Docker


Contributing
------------

* Fork the project.
* Make your feature addition or bug fix.
* Add tests for it. This is important so I don't break it in a future version unintentionally.
* Send a pull request. Bonus points for topic branches.


TODOs/Problems
--------------

* Needs documentation around swarm service


Discussion
----------

If you have questions, please make an issue.


License
-------

This project is released under the [AGPLv3 License](LICENSE).
