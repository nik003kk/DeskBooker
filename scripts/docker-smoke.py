"""Exercise the real containers; uses throwaway CI credentials, never prints tokens."""
import json
import os
import sys
import urllib.error
import urllib.request

base = "http://localhost:5173"
email = "compose-smoke@" + os.environ["REGISTRATION_EMAIL_DOMAIN"]
password = "Compose-smoke-Password-2026!"

def request(path, expected, data=None, token=None):
    headers = {}
    if data is not None:
        headers["Content-Type"] = "application/json"
    if token:
        headers["Authorization"] = "Bearer " + token
    req = urllib.request.Request(base + path, headers=headers,
        data=json.dumps(data).encode() if data is not None else None)
    try:
        with urllib.request.urlopen(req, timeout=15) as response:
            code, body = response.status, response.read()
    except urllib.error.HTTPError as error:
        code, body = error.code, error.read()
    assert code == expected, f"{path}: expected {expected}, got {code}"
    return body

assert b'<div id="root"' in request("/desks", 200)
request("/api/health/db", 200)
request("/api/resources?type=DESK", 401)
if "--existing" not in sys.argv:
    request("/api/auth/register", 201, {"email": email, "password": password,
        "first_name": "Compose", "last_name": "Smoke"})
login = json.loads(request("/api/auth/login", 200, {"email": email, "password": password}))
token = login["access_token"]
me = json.loads(request("/api/auth/me", 200, token=token))
assert me["user"]["email"] == email
resources = json.loads(request("/api/resources?type=DESK", 200, token=token))
assert isinstance(resources["resources"], list)
request("/api/admin/users", 403, token=token)
print("Docker smoke passed: SPA, proxy, database, auth and current permissions")
