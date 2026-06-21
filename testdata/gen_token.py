#!/usr/bin/env python3
"""Generate a throwaway compact IBCT for CI smoke tests.

The token is structurally valid (parseable header + section 3.1 claims)
but its issuer does not resolve and its signature is not real, so burling
reports findings. That is exactly what the smoke test wants: a token that
exercises the validator and yields a non-empty SARIF document, without
committing any signed fixture to the repository.

Usage:
    python3 testdata/gen_token.py [output-path]   # default: token.jwt
"""
import base64
import json
import sys
import time


def b64(obj):
    raw = json.dumps(obj, separators=(",", ":")).encode()
    return base64.urlsafe_b64encode(raw).rstrip(b"=").decode()


def main():
    out = sys.argv[1] if len(sys.argv) > 1 else "token.jwt"
    now = int(time.time())
    header = {"alg": "EdDSA", "typ": "aip+jwt", "kid": "k1"}
    payload = {
        "iss": "aip:web:nonexistent.invalid/agent",
        "sub": "aip:web:nonexistent.invalid/agent-bob",
        "scope": {"tools": ["search"]},
        "budget_usd": 5,
        "max_depth": 0,
        "iat": now - 60,
        "exp": now + 1800,
    }
    sig = base64.urlsafe_b64encode(b"\x00" * 64).rstrip(b"=").decode()
    with open(out, "w") as f:
        f.write(b64(header) + "." + b64(payload) + "." + sig)
    print("wrote", out)


if __name__ == "__main__":
    main()
